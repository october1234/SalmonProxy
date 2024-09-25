package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"salmonproxy/hosts"
	"salmonproxy/proxy"
)

func main() {
	localAddr := "0.0.0.0:25565"

	hosts.LoadConfig()

	go setupHttp()

	laddr, err := net.ResolveTCPAddr("tcp", localAddr)
	if err != nil {
		log.Printf("Failed to resolve local address: %s\n", err)
		os.Exit(1)
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Printf("Failed to open local port to listen: %s\n", err)
		os.Exit(1)
	}

	log.Printf("Salmon Proxy 🐟 listenting on %v \n", localAddr)

	connid := uint64(0)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Printf("Failed to accept connection '%s'\n", err)
			continue
		}
		connid++

		p := proxy.New(conn, laddr)

		go p.Start()
	}
}

func setupHttp() {
	http.HandleFunc("/config", hosts.AcceptNewConfig)
	http.HandleFunc("/config/reload", func(w http.ResponseWriter, r *http.Request) {
		hosts.LoadConfig()
		json.NewEncoder(w).Encode(map[string]interface{}{"message": "Successfully Reloaded Config"})
	})
	err := http.ListenAndServe(":80", nil)
	if err != nil {
		log.Printf("Failed to start http server: %s\n", err)
	}
}
