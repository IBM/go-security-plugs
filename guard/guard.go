package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type minmax struct {
	L float32
	H float32
}
type wsgateConfig struct {
	QsKeys []minmax
	Tbd    []minmax
	Level  int
}

func fetchConfig(w http.ResponseWriter, req *http.Request) {
	query := req.URL.Query()
	nsSlice := query["ns"]
	srvSlice := query["srv"]
	fmt.Printf("Servicing fetchConfig %v\n", query)
	if len(nsSlice) != 1 || len(srvSlice) != 1 {
		fmt.Printf("Servicing fetchConfig missing data %d %d\n", len(nsSlice), len(srvSlice))
		return
	}
	ns := nsSlice[0]
	srv := srvSlice[0]
	if ns == "" || srv == "" {
		fmt.Printf("Servicing fetchConfig missing data\n")
		return
	}
	fmt.Printf("Servicing fetchConfig of %s.%s\n", ns, srv)
	data := new(wsgateConfig)
	data.QsKeys = append(data.QsKeys, minmax{1, 5})
	data.QsKeys = append(data.QsKeys, minmax{10, 51})
	data.Tbd = append(data.Tbd, minmax{1, 5})
	data.Level = 1
	fmt.Printf("data %v\n", data)
	buf, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Servicing fetchConfig error while Marshal %v\n", err)
	}
	fmt.Printf("buf %v\n", buf)
	w.Write(buf)
}

func main() {
	http.HandleFunc("/fetchConfig", fetchConfig)
	http.ListenAndServe(":80", nil)
}
