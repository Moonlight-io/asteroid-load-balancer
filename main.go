package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
)

// Config holds the Asteriod configuration.
type Config struct {
	// List of seeds that will be used to spread the load. Hence will be included
	// in the polling algorithm.
	Seeds []string
	// Time between each network poll.
	HeartbeatInterval time.Duration
	// HTTP handler timeout
	HTTPTimeout time.Duration
	// Port Asteriod will listen on
	ListenAddr string
}

// Asteriod represents the main object that maintains a predefined list of
// nodes. Every heartbeat interval Asteriod will poll its nodes to update their
// status. Asteriod's HTTP endpoint is available on the ListenAddr defined in
// the configuration.
type Asteriod struct {
	// Config holds the Asteriod configuration.
	Config

	nodes    []*node
	topNodes []*node
}

// newAsteriod constructs a new Asteriod object.
func newAsteriod(cfg Config) *Asteriod {
	return &Asteriod{
		Config: cfg,
		nodes:  makeNodes(cfg.Seeds),
	}
}

// Poll the known nodes and update their status.
func (a *Asteriod) checkHeartbeat() {
	for i, n := range a.nodes {
		a.nodes[i].getBlockHeight()

		var newTopNodes []*node
		if len(newTopNodes) == 0 || a.nodes[i].blockHeight == newTopNodes[0].blockHeight {
			newTopNodes = append(newTopNodes, a.nodes[i])
		} else if a.nodes[i].blockHeight > newTopNodes[0].blockHeight {
			newTopNodes = []*node{a.nodes[i]}
		}
		for _, tn := range a.topNodes {
			if tn.target == n.target {
				a.nodes[i].count = a.nodes[i].count + tn.count
			}
		}
		a.topNodes = newTopNodes
	}
}

func (a *Asteriod) handleProxy(w http.ResponseWriter, r *http.Request) error {
	if len(a.topNodes) == 0 {
		return nil
	}

	i := rand.Intn(len(a.topNodes))

	//reset hostname and nuke cloudflare headers
	r.Host = a.topNodes[i].target.Hostname()
	r.Header.Del("cf-connecting-ip")
	r.Header.Del("cf-visitor")
	r.Header.Del("cf-ipcountry")
	r.Header.Del("cf-ray")

	a.topNodes[i].proxy.ServeHTTP(w, r)

	return nil
}

func (a *Asteriod) handleRegister(w http.ResponseWriter, r *http.Request) error {
	var body registrant
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return err
	}
	defer r.Body.Close()

	for index, n := range a.nodes {
		if n.target.String() == body.Target {
			a.nodes[index].state = stateActive
			return nil
		}
	}

	ip, _ := url.Parse(body.Target)
	newNode := newNode(ip, 0)
	newNode.getBlockHeight()
	a.nodes = append(a.nodes, newNode)

	return nil
}

func (a *Asteriod) handleUnregister(w http.ResponseWriter, r *http.Request) error {
	var body registrant
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return err
	}
	defer r.Body.Close()

	for i, n := range a.nodes {
		fmt.Println(n.target)
		if n.target.String() == body.Target {
			a.nodes[i] = a.nodes[len(a.nodes)-1]
			a.nodes = a.nodes[:len(a.nodes)-1]
			return nil
		}
	}
	return nil
}

func (a *Asteriod) serveHTTP() {
	r := mux.NewRouter()
	r.HandleFunc("/", makeHTTPHandler(a.handleProxy))
	r.HandleFunc("/register", makeHTTPHandler(a.handleRegister))
	r.HandleFunc("/unregister", makeHTTPHandler(a.handleUnregister))
	http.Handle("/", r)

	handler := http.TimeoutHandler(r, a.HTTPTimeout, "Timeout!")
	if err := http.ListenAndServe(a.ListenAddr, handler); err != nil {
		log.Fatal(err)
	}
}

// loop starts the a loop which will check the status of the known nodes each
// heartbeat interval.
func (a *Asteriod) loop() {
	hearbeatTimer := time.NewTimer(a.HeartbeatInterval)
	for {
		select {
		case <-hearbeatTimer.C:
			log.Println("polling nodes")
			a.checkHeartbeat()
			hearbeatTimer.Reset(a.HeartbeatInterval)
		}
	}
}

func (a *Asteriod) start() {
	// start the main loop in other routine.
	go a.loop()
	a.checkHeartbeat()
	a.serveHTTP()
}

type registrant struct {
	Target string
}

func makeNodes(seeds []string) []*node {
	nodes := make([]*node, len(seeds))
	for i, seed := range seeds {
		ip, err := url.Parse(seed)
		if err != nil {
			log.Printf("skipping seed %s due to error: %s", seed, err)
		}
		nodes[i] = newNode(ip, 0)
	}
	return nodes
}

// apiFunc is a more idiomatic way to handle HTTP requests.
type apiFunc func(w http.ResponseWriter, r *http.Request) error

// makeHTTPHandler converts a apiFunc in a compatible http.HandlerFunc.
func makeHTTPHandler(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			log.Printf("http error: %s", err)
		}
	}
}

func main() {
	config := Config{
		ListenAddr:        ":8080",
		HTTPTimeout:       2 * time.Second,
		HeartbeatInterval: 20 * time.Second,
		Seeds: []string{
			"http://seed1.cityofzion.io:8080",
			"http://seed2.cityofzion.io:8080",
			"http://seed3.cityofzion.io:8080",
			"http://seed4.cityofzion.io:8080",
			"http://pyrpc1.neeeo.org:10332",
		},
	}

	newAsteriod(config).start()
}

func init() {
	rand.Seed(time.Now().Unix())
	log.SetFlags(0)
	log.SetPrefix("[asteroid] ")
}
