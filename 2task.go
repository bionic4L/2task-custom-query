package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
)

type QueryData struct {
	name   string
	paramV string
}

var QueryDataMap = make(map[string](chan string))
var mutex = &sync.Mutex{}

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
	mutex.Lock()
	defer mutex.Unlock()
	ch := make(chan string, 3)
	ch <- v
	QueryDataMap[k] = ch
}

func ReadChan(k string) (chan string, bool) {
	mutex.Lock()
	defer mutex.Unlock()
	ch, ok := QueryDataMap[k]
	return ch, ok
}

func handlePUT(w http.ResponseWriter, r *http.Request) {
	statusBadReq := "\nStatus code: " + strconv.Itoa(http.StatusBadRequest) + " Bad Request\n"
	statusOK := "\nStatus code: " + strconv.Itoa(http.StatusOK) + " OK\n"

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
	statusNotFound := "\nStatus code: " + strconv.Itoa(http.StatusNotFound) + " Not Found\n"

	urlRaw := r.URL
	incomeQueryName, _ := ParseQuery(urlRaw)

	ch, ok := ReadChan(incomeQueryName)

	if !ok {
		w.Write([]byte(statusNotFound))
	} else {
		recievedStr := <-ch
		w.Write([]byte(recievedStr))
	}

}
