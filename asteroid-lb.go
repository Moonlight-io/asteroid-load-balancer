package main

import (
"net/http"
"net/url"
"net/http/httputil"
"encoding/json"
"fmt"
"github.com/gorilla/mux"
"math/rand"
"time"
)


type registrant struct {
	Target string
}


type node struct {
	target *url.URL
	blockHeight int
	state string
	proxy *httputil.ReverseProxy
}


var nodes []node

var topNodes []node

func heartbeat(){

	for {
		var newTopNodes []node

		for index, n := range nodes {
			nodes[index].blockHeight = getBlockHeight(n)

			if (len(newTopNodes) == 0) || (nodes[index].blockHeight == newTopNodes[0].blockHeight){
				newTopNodes = append(newTopNodes, nodes[index])
			} else if nodes[index].blockHeight > newTopNodes[0].blockHeight {
				newTopNodes = []node{nodes[index]}
			}
		}

		topNodes = newTopNodes
		time.Sleep(5 * time.Second)
	}
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

func proxy(w http.ResponseWriter, r *http.Request) {
	n := topNodes[rand.Intn(len(topNodes))]
	n.proxy.ServeHTTP(w,r)
}


func main() {

	rand.Seed(time.Now().Unix())

	go heartbeat()

	r := mux.NewRouter()
	r.HandleFunc("/reg", register)     //protect to prevent MitM
	r.HandleFunc("/dereg", deregister) //protect to prevent MitM
	r.HandleFunc("/", proxy)
	http.Handle("/", r)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}

}