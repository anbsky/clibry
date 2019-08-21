package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

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
}

type QueryStream struct {
	queries []*Query
	length  int
}

func LaunchClients(n int) {
	wg := sync.WaitGroup{}
	lqs := parseLog("app.log")
	stream := createStream(lqs)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(n int, wg *sync.WaitGroup) {
			fmt.Printf("starting stream #%v of %v queries\n", n, stream.Length())
			stream.playback()
			fmt.Printf("done stream #%v\n", n)
			wg.Done()
		}(n, &wg)
	}
	wg.Wait()
}

func (s *QueryStream) playback() {
	// var queries []*Query
	for _, q := range s.queries {
		q.send()
	}
}

func (s *QueryStream) Length() int {
	return s.length
}

func createStream(lqs <-chan LoggedQuery) *QueryStream {
	var queries []*Query
	for lq := range lqs {
		q := &Query{lq.Method, lq.Params, ""}
		queries = append(queries, q)
	}
	return &QueryStream{queries, len(queries)}
}

func parseLog(fileName string) chan LoggedQuery {
	ret := make(chan LoggedQuery)
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}

	s := bufio.NewScanner(f)
	fmt.Println(s.Err())
	go func() {
		for s.Scan() {
			var q LoggedQuery
			err := json.Unmarshal(s.Bytes(), &q)
			if err == nil {
				ret <- q
			}
		}
		fmt.Println(s.Err())
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
