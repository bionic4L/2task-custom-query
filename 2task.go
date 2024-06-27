package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"sync"
)

type Query struct {
	sync.RWMutex
	QueryDataMap map[string](chan string)
}

var q = &Query{
	QueryDataMap: make(map[string](chan string)),
}

var statusBadReq = fmt.Sprintf("\nStatus code: %v Bad Request\n", http.StatusBadRequest)
var statusOK = fmt.Sprintf("\nStatus code: %v OK\n", http.StatusOK)
var statusNotFound = fmt.Sprintf("\nStatus code: %v Not Found\n", http.StatusNotFound)

func main() {

	portFlag := flag.String(
		"port",
		":8080", //default value if extra flag not specified
		"adding your listening port",
	)
	flag.Parse()

	http.HandleFunc("/", handleRequest)

	err := http.ListenAndServe(*portFlag, nil)
	if err != nil {
		err = fmt.Errorf("Can't start server. Error: %v", err)
	}
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
	parsedName := url.Path[1:]
	parsedParamValue := url.Query().Get("v")

	return parsedName, parsedParamValue
}

func AddChan(k string, v string) {
	queryMap := q.QueryDataMap
	defer q.Unlock()
	_, ok := queryMap[k]
	if !ok {
		q.Lock()
		ch := make(chan string, 3)
		ch <- v
		queryMap[k] = ch
	} else {
		q.Lock()
		queryMap[k] <- v
	}

}

func ReadChan(k string) (chan string, bool) {
	q.RLock()
	queryMap := q.QueryDataMap

	ch, ok := queryMap[k]
	q.RUnlock()

	return ch, ok
}

func handlePUT(w http.ResponseWriter, r *http.Request) {

	urlRaw := r.URL
	queryName, paramValue := ParseQuery(urlRaw)

	if paramValue == "" {
		w.Write([]byte(statusBadReq))
	} else {
		AddChan(queryName, paramValue)
		w.Write([]byte(statusOK))
	}
}
func handleGET(w http.ResponseWriter, r *http.Request) {
	urlRaw := r.URL
	incomeQueryName, _ := ParseQuery(urlRaw)

	ch, ok := ReadChan(incomeQueryName)

	if !ok {
		w.Write([]byte(statusNotFound))
	} else {
		recievedStr := <-ch
		if recievedStr == "" {
			w.Write([]byte(statusNotFound))
		}
		w.Write([]byte(recievedStr))
	}

}
