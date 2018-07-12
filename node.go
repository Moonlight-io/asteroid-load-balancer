package main

import (
	"encoding/json"
	"net/http"
	"time"
	"bytes"
	"net/url"
	"net/http/httputil"
)

type node struct {
	target *url.URL
	blockHeight int
	state string
	proxy *httputil.ReverseProxy
}

func getBlockHeight(Node node) int {
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"method": "getblockcount",
		"params": []string{},
		"id": "0"}

	bodyBytes, _ := json.Marshal(body)

	client := http.Client{Timeout: time.Duration(2 * time.Second)}

	resp, err := client.Post(Node.target.String(), "application/json", bytes.NewBuffer(bodyBytes))
	if nil != err {
		return -1
	}

	var res map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&res)

	return int(res["result"].(float64))
}
