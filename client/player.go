package client

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptrace"
	"sync"
	"time"

	ljsonrpc "github.com/lbryio/lbry.go/extras/jsonrpc"
)

const streamPrefix = "http://localhost:8080/content/claims"

type Stream struct {
	ClaimID         string
	Name            string
	SizeBytes       uint64
	RetrievedBytes  int
	TimeToFirstByte time.Duration
	TotalTime       time.Duration
}

type StreamBatchStats struct {
	Successes uint
	Failures  uint
	Elapsed   time.Duration
}

type StreamBatch struct {
	streams []*Stream
	length  int
	Stats   *StreamBatchStats
}

func LaunchStreams(n int) {
	wg := sync.WaitGroup{}

	claims := getHomepageIDs()

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(n int) {
			batch := &StreamBatch{}
			batch.fill(claims)
			fmt.Printf("starting playback #%v\n", n)
			batch.run()
			fmt.Printf("done stream #%v\n", n)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func getHomepageIDs() []ljsonrpc.Claim {
	var claimList ljsonrpc.ClaimListResponse
	q := &Query{
		Method: "claim_search",
		Params: ClaimSearchParams{
			PageSize: 20, Page: 1, NoTotals: true,
			AnyTags: []string{"blockchain", "comedy", "economics", "gaming"},
			NotTags: []string{"porn", "nsfw", "mature", "xxx"},
			OrderBy: []string{"trending_global", "trending_mixed"},
		},
	}
	r, _, err := q.send()
	if err != nil {
		log.Fatal(err)
	}

	if err = ljsonrpc.Decode(r.Result, &claimList); err != nil {
		log.Fatal(err)
	}

	return claimList.Claims
}

func (b *StreamBatch) fill(claims []ljsonrpc.Claim) {
	var streams []*Stream

	rand.Seed(time.Now().Unix())

	for _, c := range claims {
		var size uint64
		s := c.Value.GetStream()
		if s != nil {
			size = s.Source.GetSize()
		}
		streams = append(streams, &Stream{ClaimID: c.ClaimID, Name: c.Name, SizeBytes: size})
	}
	rand.Shuffle(len(streams), func(i, j int) { streams[i], streams[j] = streams[j], streams[i] })
	b.streams = streams
}

func (b *StreamBatch) run() {
	b.Stats = &StreamBatchStats{}
	start := time.Now()
	for _, s := range b.streams {
		fmt.Printf("playing %v (%v bytes)\n", s.URL(), s.SizeBytes)
		s.run()
		fmt.Printf("%v bytes played out of %v, ttfb: %.4f, total: %.4f\n", s.RetrievedBytes, s.SizeBytes, s.TimeToFirstByte.Seconds(), s.TotalTime.Seconds())
	}
	b.Stats.Elapsed = time.Since(start)
}

func (s *Stream) run() {
	req, _ := http.NewRequest("GET", s.URL(), nil)
	playBytes := float64(s.SizeBytes) * .25
	if playBytes > 0 {
		req.Header.Add("Range", fmt.Sprintf("bytes=0-%v", uint64(playBytes)))
	}
	var start time.Time
	trace := &httptrace.ClientTrace{
		GotFirstResponseByte: func() {
			s.TimeToFirstByte = time.Since(start)
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	start = time.Now()
	r, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		log.Print(err)
	}
	defer r.Body.Close()
	buf := make([]byte, 4096)
	var ret int
	for {
		n, err := r.Body.Read(buf)
		if err != nil {
			break
		}
		ret += n
	}
	s.RetrievedBytes = ret
	s.TotalTime = time.Since(start)
}

func (s *Stream) URL() string {
	return fmt.Sprintf("%v/%v/%v/stream", streamPrefix, s.Name, s.ClaimID)
}
