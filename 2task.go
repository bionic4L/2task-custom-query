package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

type Query struct {
	sync.RWMutex
	QueryDataMap map[string](chan string)
}

var q = &Query{
	QueryDataMap: make(map[string](chan string)),
}

var (
	statusBadReq   = fmt.Sprintf("\nStatus code: %v Bad Request\n", http.StatusBadRequest)
	statusOK       = fmt.Sprintf("\nStatus code: %v OK\n", http.StatusOK)
	statusNotFound = fmt.Sprintf("\nStatus code: %v Not Found\n", http.StatusNotFound)
)

func main() {
	portFlag := flag.String(
		"port",
		":8080", //default value if extra flag not specified
		"adding your listening port",
	)
	flag.Parse()

	http.HandleFunc("/", handleRequest)

	log.Fatal(http.ListenAndServe(*portFlag, nil))
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPut:
		handlePUT(w, r)
	case http.MethodGet:
		handleGET(w, r)
	default:
		http.Error(w, "Method is not supported!", http.StatusMethodNotAllowed)
	}
}

func ParseQuery(url *url.URL) (qName string, qValue string) {

	return url.Path[1:], url.Query().Get("v")
}

func AddChan(k string, v string) {
	q.Lock()
	defer q.Unlock()

	_, ok := q.QueryDataMap[k]

	if ok {
		q.QueryDataMap[k] <- v
		return
	}

	ch := make(chan string, 3)
	ch <- v
	q.QueryDataMap[k] = ch

}

func ReadChan(ctx context.Context, k string) (chan string, bool) {
	q.RLock()
	defer q.RUnlock()
	time.Sleep(5 * time.Second)
	ch, ok := q.QueryDataMap[k]

	return ch, ok

}

func handlePUT(w http.ResponseWriter, r *http.Request) {
	urlRaw := r.URL
	queryName, paramValue := ParseQuery(urlRaw)

	if paramValue == "" {
		w.Write([]byte(statusBadReq))
		return
	}

	AddChan(queryName, paramValue)
	w.Write([]byte(statusOK))
}
func handleGET(w http.ResponseWriter, r *http.Request) {
	urlRaw := r.URL
	n, _ := strconv.Atoi(urlRaw.Query().Get("timeout"))
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(n)*time.Second)
	defer cancel()

	incomeQueryName, _ := ParseQuery(urlRaw)

	ch, _ := ReadChan(ctx, incomeQueryName)

	select {
	case <-ctx.Done():
		w.Write([]byte(statusNotFound))

	case rec := <-ch:
		w.Write([]byte(rec))
	}

}
