package main

import (
	"encoding/json"
	"net/http"
	"time"
	"bytes"
	"fmt"
	"net/url"
	"net/http/httputil"
)

type node struct {
	target *url.URL
	blockHeight int
	state string
	proxy *httputil.ReverseProxy
	count int
}

func getBlockHeight(Node node) int {
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"method": "getblockcount",
		"params": []string{},
		"id": "0"}

	bodyBytes, _ := json.Marshal(body)

	client := http.Client{Timeout: time.Duration(6 * time.Second)}

	resp, err := client.Post(Node.target.String(), "application/json", bytes.NewBuffer(bodyBytes))
	if nil != err {
		fmt.Println(fmt.Sprintf("%s  : dead",Node.target))
		fmt.Println(err)
		return -1
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	if res == nil {
		fmt.Println(fmt.Sprintf("%s  : dead 2",Node.target))
		return -1
	}
	fmt.Println(fmt.Sprintf("%s  : %d",Node.target, int(res["result"].(float64))))
	return int(res["result"].(float64))
}
