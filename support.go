package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type registrant struct {
	Target string
}

func register(w http.ResponseWriter, r *http.Request) error {
	var body registrant
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return err
	}
	defer r.Body.Close()

	for index, n := range nodes {
		if n.target.String() == body.Target {
			nodes[index].state = stateActive
			return nil
		}
	}

	ip, _ := url.Parse(body.Target)
	newNode := newNode(ip, 0)
	newNode.getBlockHeight()
	nodes = append(nodes, newNode)

	return nil
}

func deregister(w http.ResponseWriter, r *http.Request) error {
	var body registrant
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return err
	}
	defer r.Body.Close()

	for i, n := range nodes {
		fmt.Println(n.target)
		if n.target.String() == body.Target {
			nodes[i] = nodes[len(nodes)-1]
			nodes = nodes[:len(nodes)-1]
			return nil
		}
	}
	return nil
}
