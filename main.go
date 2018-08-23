package main

import (
	"github.com/gorilla/mux"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"
)

const heartbeatInterval = 20 * time.Second

var (
	nodes    []*node
	topNodes []*node
)

func beat() {
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
}

func proxy(w http.ResponseWriter, r *http.Request) {
	if len(topNodes) > 0 {
		i := rand.Intn(len(topNodes))

		//reset hostname and nuke cloudflare headers
		r.Host = topNodes[i].target.Hostname()
		r.Header.Del("cf-connecting-ip")
		r.Header.Del("cf-visitor")
		r.Header.Del("cf-ipcountry")
		r.Header.Del("cf-ray")

		topNodes[i].proxy.ServeHTTP(w, r)
	}
}

func main() {
	//define your seeds
        seeds := []string{
                "https://seed1.cityofzion.io:443",
                "https://seed2.cityofzion.io:443",
                "https://seed3.cityofzion.io:443",
                "https://seed4.cityofzion.io:443",
                "https://seed5.cityofzion.io:443",
                "https://seed6.cityofzion.io:443",
                "https://seed7.cityofzion.io:443",
                "https://seed8.cityofzion.io:443",
                "https://seed9.cityofzion.io:443",
                "https://seed0.cityofzion.io:443",
        }

	nodes = make([]*node, len(seeds))
	for i, n := range seeds {
		ip, err := url.Parse(n)
		if err != nil {
			log.Fatal(err)
		}
		node := newNode(ip, 0)
		nodes[i] = node
	}

	//setup the node heartbeat
	go func() {
		beat()
		heartbeatTimer := time.NewTicker(heartbeatInterval)
		for {
			select {
			case <-heartbeatTimer.C:
				beat()
				time.Sleep(heartbeatInterval)
			}
		}
	}()

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
