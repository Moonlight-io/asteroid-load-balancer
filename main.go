package main

import (
"net/http"
"github.com/gorilla/mux"
"math/rand"
"time"
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
		}

		topNodes = newTopNodes
		time.Sleep(5 * time.Second)
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