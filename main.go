package main

import (
	dsm "dsm/shmem"
	"io"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"strconv"
)

var sm *dsm.DSM

func getRoot(w http.ResponseWriter, r *http.Request) {
	log.Printf("got / request\n")
	sm.Set("Hello", 0)
	sm.Set("Speed", 0)
	io.WriteString(w, "This is my website!\n")
}

func getHello(w http.ResponseWriter, r *http.Request) {
	t := sm.Get("Hello")
	sm.Set("Hello", (*t).(int)+1)
	log.Printf("got /hello request no %v\n", *t)
	io.WriteString(w, "Hello, HTTP!\n")
}

func getSpeed(w http.ResponseWriter, r *http.Request) {
	t := sm.Get("Speed")
	log.Printf("got /speed request no %v\n", *t)
	io.WriteString(w, "Speed, HTTP!"+strconv.Itoa((*t).(int)))
}

func increaseSpeed(w http.ResponseWriter, r *http.Request) {
	t := sm.Get("Speed")
	sm.Set("Speed", (*t).(int)+1)
	// sm.SendWriteUpdate(dsm.SetArgs{Name: "Speed", Value: (*t).(int) + 1, Creds: dsm.Creds{SenderId: sm.Id}})
	log.Printf("got /speed request no %v\n", *t)
	io.WriteString(w, "Speed, HTTP!"+strconv.Itoa((*t).(int)))
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
	http.HandleFunc("/speed", getSpeed)
	http.HandleFunc("/increaseSpeed", increaseSpeed)

	sm = new(dsm.DSM)

	rpc.Register(sm)
	rpc.HandleHTTP()

	l, err := net.Listen("tcp", rpcPort)
	if err != nil {
		log.Panicf("listen error: %s\n", err)
	}
	log.Printf("[INFO] Serving RPC server on port %s\n", rpcPort)
	go http.Serve(l, nil)

	sm.Init(registryAddr, "localhost"+rpcPort)
	log.Printf("[INFO] Serving HTTP server on port %s\n", serverPort)
	http.ListenAndServe(serverPort, nil)
}
