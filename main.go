package main

import (
	"os"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"dsm/shmem"
)

var sm *dsm.DSM

func getRoot(w http.ResponseWriter, r *http.Request) {
	log.Printf("got / request\n")
	io.WriteString(w, "This is my website!\n")
}
func getHello(w http.ResponseWriter, r *http.Request) {
	log.Printf("got /hello request\n")
	// sm.SendBroadcast()
	io.WriteString(w, "Hello, HTTP!\n")
}

func main() {
	// Take args from cmd line
	// 1 - server port
	// 2 - rpc port
	// 3 - registry address
	if len(os.Args) < 3 {
		log.Fatalf("Usage: %s <server port> <rpc port> <registry address>\n", os.Args[0])
		return
	}

	serverPort := os.Args[1]
	rpcPort := os.Args[2]
	if serverPort == rpcPort {
		log.Fatalf("Server port and RPC port must be different\n")
		return
	}

	registryAddr := ""
	if len(os.Args) == 4 {
		registryAddr = os.Args[3]
	} else {
		log.Printf("[INFO] Running as registry\n")
	}


	http.HandleFunc("/", getRoot)
	http.HandleFunc("/hello", getHello)
	sm = new(dsm.DSM)
	sm.Init(registryAddr, "localhost"+rpcPort)

	rpc.Register(sm)
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", rpcPort)
	if err != nil {
		log.Panicf("listen error: %s\n", err)
	}
	log.Printf("[INFO] Serving RPC server on port %s\n", rpcPort)
	go http.Serve(l, nil)

	log.Printf("[INFO] Serving HTTP server on port %s\n", serverPort)
	http.ListenAndServe(serverPort, nil)
}