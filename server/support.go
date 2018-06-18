package server

import (
	"net/http"
	"encoding/json"
	"fmt"
	"net/url"
	"net/http/httputil"
)

type registrant struct {
	Target string
}

func register(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var body registrant
	decoder.Decode(&body)
	defer r.Body.Close()

	for index, n := range nodes {
		if n.target.String() == body.Target {
			nodes[index].state = "active"
			return
		}
	}
	ip, _ := url.Parse(body.Target)

	newNode := node{target: ip, blockHeight: 0, state: "active", proxy: httputil.NewSingleHostReverseProxy(ip)}
	newNode.blockHeight = getBlockHeight(newNode)
	nodes = append(nodes, newNode)
}

func deregister(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)

	var body registrant
	decoder.Decode(&body)
	defer r.Body.Close()

	for i, n := range nodes {
		fmt.Println(n.target)
		if n.target.String() == body.Target {
			nodes[i] = nodes[len(nodes) - 1 ]
			nodes = nodes[:len(nodes) - 1]
			return
		}
	}
}


