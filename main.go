package main

import (
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
)

const heartbeatInterval = 20 * time.Second

var (
	nodes    []*node
	topNodes []*node
)

func heartbeat() {
	heartbeatTimer := time.NewTicker(heartbeatInterval)
	for {
		select {
		case <-heartbeatTimer.C:
			log.Println("polling nodes")

			var newTopNodes []*node
			for index, n := range nodes {
				nodes[index].getBlockHeight()
				if (len(newTopNodes) == 0) || (nodes[index].blockHeight == newTopNodes[0].blockHeight) {
					newTopNodes = append(newTopNodes, nodes[index])
				} else if nodes[index].blockHeight > newTopNodes[0].blockHeight {
					newTopNodes = []*node{nodes[index]}
				}

				for _, tn := range topNodes {
					if tn.target == n.target {
						nodes[index].count = nodes[index].count + tn.count
					}
				}
			}
			topNodes = newTopNodes
			time.Sleep(heartbeatInterval)
		}
	}
}

func proxy(w http.ResponseWriter, r *http.Request) {
	if len(topNodes) > 0 {
		i := rand.Intn(len(topNodes))
		topNodes[i].count = topNodes[i].count + 1
		topNodes[i].proxy.ServeHTTP(w, r)
	}
}

func main() {
	//define your seeds
	seeds := []string{
		"http://seed1.cityofzion.io:8080",
		"http://seed2.cityofzion.io:8080",
		"http://seed3.cityofzion.io:8080",
		"http://seed4.cityofzion.io:8080",
		"http://pyrpc1.neeeo.org:10332",
	}

	nodes := make([]*node, len(seeds))
	for i, n := range seeds {
		ip, err := url.Parse(n)
		if err != nil {
			log.Fatal(err)
		}
		node := newNode(ip, 0)
		if err := node.getBlockHeight(); err != nil {
			log.Println(err)
			// skipping bad nodes? Or should we add statusDead?
			continue
		}
		nodes[i] = node
	}

	go heartbeat()

	r := mux.NewRouter()

	//Enable if you dont mind MaMs
	//r.HandleFunc("/reg", register)     //protect to prevent MitM
	//r.HandleFunc("/dereg", deregister) //protect to prevent MitM
	r.HandleFunc("/", proxy)
	http.Handle("/", r)

	handler := http.TimeoutHandler(r, time.Second*2, "Timeout!")

	if err := http.ListenAndServe(":8080", handler); err != nil {
		log.Fatal(err)
	}
}

func init() {
	rand.Seed(time.Now().Unix())
	log.SetFlags(0)
	log.SetPrefix("[asteroid] ")
}
