package dsm

import (
	"log"
	"sync"
	"net/rpc"
)



type DSM struct {
	Id int // Unique ID for the client
	addrs []string // Addresses of all clients
	addrLock sync.Mutex 
	vars map[string]*interface{} // Shared variables
}

// Initialize the DSM
func (d *DSM) Init(serverAddress string, myAddress string) error {
	d.vars = make(map[string]*interface{})
	if serverAddress == "" {
		// Start as the registry service to discover clients
		d.Id = 0
		d.addrs = []string{myAddress}
		return nil
	} else {
		// Register with the registry service
		client, err := rpc.DialHTTP("tcp", serverAddress)
		if err != nil {
			return err
		}
		var reply RegisterReply
		err = client.Call("DSM.RegisterReplica", &myAddress, &reply)
		if err != nil {
			return err
		}
		d.Id = reply.Id
		log.Printf("Got addrs: %v\n", d.addrs)
	}
	return nil
}

func (d *DSM) Get(name string) *interface{} {
	// Get the pointer to a shared variable
	// This returns nil if the variable does not exist
	return d.vars[name]
}

func (d *DSM) Set(name string, value interface{}) {
	// Set the value of a shared variable
	if d.vars[name] == nil {
		d.vars[name] = &value
	} else {
		*d.vars[name] = value
	}
}

func (d *DSM) SendBroadcast(fun string, args interface{}) []int {
	log.Printf("[DEBUG] Sending broadcast to %s\n", d.addrs)
	replies := make([]int, 0)
	for _, addr := range d.addrs {
		client, err := rpc.DialHTTP("tcp", addr)
		if err != nil {
			log.Printf("Error dialing %s: %v\n", addr, err)
			continue
		}
		var reply int
		log.Printf("[DEBUG] Calling %s(%s) on %s\n",fun, args, addr)
		err = client.Call(fun, args, &reply)
		if err != nil {
			log.Printf("Error calling %s: %v\n", addr, err)
			continue
		}
		replies = append(replies, reply)
	}
	log.Printf("[DEBUG] Got replies: %v\n", replies)
	return replies
}

