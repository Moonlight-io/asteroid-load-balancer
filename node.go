package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

// state represents the state of a remote node.
type state uint8

// Vialable state constants.
const (
	stateActive state = iota
)

type node struct {
	target      *url.URL
	blockHeight int
	state       state
	proxy       *httputil.ReverseProxy
	count       int
}

// newNode returns a new node object.
func newNode(ip *url.URL, height int) *node {
	return &node{
		target:      ip,
		blockHeight: height,
		state:       stateActive,
		proxy:       httputil.NewSingleHostReverseProxy(ip),
		count:       0,
	}
}

func (n *node) getBlockHeight() error {
	body := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "getblockcount",
		"params":  []string{},
		"id":      "0"}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	client := http.Client{Timeout: 6 * time.Second}

	resp, err := client.Post(n.target.String(), "application/json", bytes.NewBuffer(bodyBytes))
	if nil != err {
		return fmt.Errorf("%s : dead", n.target)
	}

	var res map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return err
	}
	if res == nil {
		return fmt.Errorf("%s  : dead 2", n.target)
	}

	log.Println(fmt.Sprintf("%s  : %d", n.target, int(res["result"].(float64))))

	n.blockHeight = int(res["result"].(float64))
	return nil
}
