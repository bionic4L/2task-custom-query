package main

import (
	"container/list"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type QueryData struct {
	name   string
	paramV string
}

var queryDataList *list.List

func main() {
	queryDataList = list.New()

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

func Push(elem interface{}, queue *list.List) {
	queue.PushBack(elem)
}

func Pop(queue *list.List) interface{} {
	return queue.Remove(queue.Front())
}

func ParseQuery(url *url.URL) (qName string, qValue string) {
	parsedName := url.Path[1:]
	parsedParamValue := url.Query().Get("v")

	return parsedName, parsedParamValue
}

func handlePUT(w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("This is PUT request\n")) //for temp check

	statusBadReq := "\nStatus code: " + strconv.Itoa(http.StatusBadRequest) + " Bad Request\n"
	statusOK := "\nStatus code: " + strconv.Itoa(http.StatusOK) + " OK\n"

	urlRaw := r.URL
	queryName, paramValue := ParseQuery(urlRaw)
	// w.Write([]byte("Query name:"))    //for temp check
	// w.Write([]byte(queryName))        //for temp check
	// w.Write([]byte("\nParam value:")) //for temp check
	// w.Write([]byte(paramValue))       //for temp check

	queryDataUnit := &QueryData{
		name:   queryName,
		paramV: paramValue,
	}

	if paramValue == "" {
		w.Write([]byte(statusBadReq))
	} else {
		Push(*queryDataUnit, queryDataList)
		// fmt.Println("In list:")                                  //for temp check
		// for e := queryDataList.Front(); e != nil; e = e.Next() { //for temp check
		// 	fmt.Println(e.Value) //for temp check
		// } //for temp check
		w.Write([]byte(statusOK))
	}
}

func handleGET(w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("This is GET request\n")) //for temp check

	statusNotFound := "\nStatus code: " + strconv.Itoa(http.StatusNotFound) + " Not Found\n"
	if queryDataList.Front() != nil {
		firstElem := queryDataList.Front().Value.(QueryData)
		qNameHave := firstElem.name
		qValueHave := firstElem.paramV

		urlRaw := r.URL

		incomeQueryName, _ := ParseQuery(urlRaw)

		if qNameHave == incomeQueryName {
			w.Write([]byte(qValueHave))
			Pop(queryDataList)
		} else {
			w.Write([]byte(statusNotFound))
		}

		// fmt.Println("Front elem: ", firstElem)   //for temp check
		// fmt.Println("Name: ", qNameHave)         //for temp check
		// fmt.Println("Param value: ", qValueHave) //for temp check
	} else {
		w.Write([]byte(statusNotFound))
	}

}
