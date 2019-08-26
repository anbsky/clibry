package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	ljsonrpc "github.com/lbryio/lbry.go/extras/jsonrpc"
	"github.com/ybbus/jsonrpc"
)

// const endpoint = "https://api.lbry.tv/api/proxy"
const endpoint = "http://localhost:8080/api/proxy"
const TokenHeader string = "X-Lbry-Auth-Token"

type LoggedQuery struct {
	Time     string      `json:"time"`
	ExecTime float64     `json:"exec_time"`
	Method   string      `json:"method"`
	Module   string      `json:"module"`
	Params   interface{} `json:"params"`
}

type Query struct {
	Method string
	Params interface{}
	Token  string
	Delay  time.Duration
}

type RPCFailure struct {
	Query    *Query
	Response *jsonrpc.RPCResponse
}

type BatchStats struct {
	Successes       uint
	Failures        uint
	RPCFailures     uint
	RPCFailuresList []*RPCFailure
	Elapsed         float64
	AverageResponse float64
}

type QueryBatch struct {
	queries []*Query
	length  int
	Stats   *BatchStats
}

func LaunchClients(n int) {
	wg := sync.WaitGroup{}
	lqs := parseLog("app.log")
	batch := createBatch(lqs)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(n int, wg *sync.WaitGroup) {
			fmt.Printf("starting batch #%v of %v queries\n", n, batch.Length())
			batch.run()
			fmt.Printf("done batch #%v (%v)\n", n, batch.Stats)
			if batch.Stats.RPCFailuresList != nil {
				for _, e := range batch.Stats.RPCFailuresList {
					fmt.Printf("[RPC error] %v\n", e)
				}
			}
			wg.Done()
		}(i, &wg)
	}
	wg.Wait()
}

func LaunchQuery(n int, url string) {
	wg := sync.WaitGroup{}

	q := &Query{Method: "resolve", Params: map[string]string{"urls": url}}
	var ecs, qNum int
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(n int, wg *sync.WaitGroup) {
			var resolveResponse ljsonrpc.ResolveResponse
			r, _, err := q.send()
			if err = ljsonrpc.Decode(r.Result, &resolveResponse); err != nil {
				log.Fatal(err)
			}
			qNum++
			if resolveResponse[url].CanonicalURL == "" {
				ecs++
			}
			wg.Done()
		}(i, &wg)
	}
	wg.Wait()
	fmt.Printf("total empty canonical urls: %v\n", ecs)
}

func (s *QueryBatch) run() {
	var rpcFailures []*RPCFailure
	start := time.Now()
	stats := &BatchStats{}
	s.Stats = stats
	for n, q := range s.queries {
		r, _, err := q.send()
		if err != nil {
			stats.Failures++
		} else {
			if r.Error != nil {
				stats.RPCFailures++
				rpcFailures = append(rpcFailures, &RPCFailure{q, r})
			} else {
				stats.Successes++
			}
		}
		stats.Elapsed = time.Since(start).Seconds()
		stats.AverageResponse = stats.Elapsed / float64(n+1)
	}
	stats.RPCFailuresList = rpcFailures
}

func (s *QueryBatch) Length() int {
	return s.length
}

func (f *RPCFailure) String() string {
	return fmt.Sprintf(
		"query: %v, response: %v", f.Query, f.Response,
	)
}

func (st BatchStats) String() string {
	return fmt.Sprintf(
		"%v successful and %v failed queries (%v RPC errors), averaged %.2f sec in a total of %.2f sec",
		st.Successes, st.Failures, st.RPCFailures, st.AverageResponse, st.Elapsed)
}

func createBatch(lqs <-chan LoggedQuery) *QueryBatch {
	var queries []*Query
	for lq := range lqs {
		q := &Query{lq.Method, lq.Params, "", 0}
		if q.Method == "" {
			fmt.Println(lq)
		}
		queries = append(queries, q)
	}
	return &QueryBatch{queries: queries, length: len(queries)}
}

func parseLog(fileName string) chan LoggedQuery {
	ret := make(chan LoggedQuery)
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}

	s := bufio.NewScanner(f)
	go func() {
		for s.Scan() {
			var q LoggedQuery
			err := json.Unmarshal(s.Bytes(), &q)
			if err == nil && q.Method != "" {
				ret <- q
			}
		}
		f.Close()
		close(ret)
	}()
	return ret
}

func (q *Query) send() (*jsonrpc.RPCResponse, time.Duration, error) {
	opts := &jsonrpc.RPCClientOpts{}
	if q.Token != "" {
		opts.CustomHeaders = map[string]string{TokenHeader: q.Token}

	}
	cl := jsonrpc.NewClientWithOpts(endpoint, opts)
	start := time.Now()
	res, err := cl.Call(q.Method, q.Params)
	return res, time.Since(start), err
}
