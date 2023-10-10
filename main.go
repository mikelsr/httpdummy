package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	hits  *sync.Map
	total *atomic.Int64
	repts *atomic.Int64
}

func main() {
	s := Server{}
	s.reset(nil, nil)
	go func() {
		for {
			s.printStats()
			time.Sleep(5 * time.Second)
		}
	}()
	http.HandleFunc("/reset/", s.reset)
	http.HandleFunc("/hit/", s.hit)
	http.ListenAndServe(":3333", nil)
}

func (s *Server) printStats() {
	fmt.Printf(`{"time": %d, "hits": %d, "repts": %d},%s`,
		time.Now().Unix(), s.total.Load(), s.repts.Load(), "\n")
}

func (s *Server) reset(_ http.ResponseWriter, _ *http.Request) {
	s.hits = &sync.Map{}
	s.total = &atomic.Int64{}
	s.repts = &atomic.Int64{}
	log.Default().Println("reset")
}

func (s *Server) hit(w http.ResponseWriter, r *http.Request) {
	s.total.Add(1)
	ids := strings.TrimPrefix(r.URL.Path, "/hit/")
	id, err := strconv.ParseInt(ids, 10, 64)
	if err != nil {
		log.Default().Println(err)
	}
	s.regHit(id)
	io.WriteString(w, htmlWithHrefs(id))
}

func (s *Server) regHit(id int64) {
	// TODO replace with counter per id so we can see repeat evolution over time
	_, repeated := s.hits.LoadOrStore(id, true)
	if repeated {
		s.repts.Add(1)
	}
}

func htmlWithHrefs(id int64) string {
	sb := strings.Builder{}
	sb.WriteString("<html><head><title>Title</title></head><body>\n")
	refCount := id % 7
	for i := int64(1); i < refCount+1; i++ {
		// nstr := strconv.FormatInt(id+i, 10)
		nstr := strconv.FormatInt(rand.Int63(), 10)
		href := fmt.Sprintf(`<p><a href="/hit/%s">%s</a></p>%s`, nstr, nstr, "\n")
		sb.WriteString(href)
	}
	sb.WriteString("</body></html>")
	return sb.String()
}
