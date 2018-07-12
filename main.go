package main

import (
	"net/http"
	"net/http/httputil"
	"github.com/gorilla/mux"
	"math/rand"
	"time"
	"net/url"
)



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

			for _, tn := range topNodes {
				if tn.target == n.target{
					nodes[index].count = nodes[index].count + tn.count
				}
			}
		}
		topNodes = newTopNodes

		time.Sleep(20 * time.Second)
	}
}

func proxy(w http.ResponseWriter, r *http.Request) {
	if len(topNodes) > 0 {
		i := rand.Intn(len(topNodes))
		topNodes[i].count = topNodes[i].count + 1
		topNodes[i].proxy.ServeHTTP(w,r)
	}
}

func main() {

	rand.Seed(time.Now().Unix())

	//define your seeds
	seeds := [...]string{"http://seed1.cityofzion.io:8080", "http://seed2.cityofzion.io:8080", "http://seed3.cityofzion.io:8080", "http://seed4.cityofzion.io:8080", "http://pyrpc1.neeeo.org:10332"}
	for _, n := range seeds {
		ip, _ := url.Parse(n)
		newNode := node{target: ip, blockHeight: 0, state: "active", proxy: httputil.NewSingleHostReverseProxy(ip), count: 0}
		newNode.blockHeight = getBlockHeight(newNode)
		nodes = append(nodes, newNode)
	}


	go heartbeat()

	r := mux.NewRouter()

	//Enable if you dont mind MaMs
	//r.HandleFunc("/reg", register)     //protect to prevent MitM
	//r.HandleFunc("/dereg", deregister) //protect to prevent MitM
	r.HandleFunc("/", proxy)
	http.Handle("/", r)

	handler := http.TimeoutHandler(r, time.Second * 2, "Timeout!")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		panic(err)
	}

}
