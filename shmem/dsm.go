package dsm

import (
	"log"
	"sync"
	"net/rpc"
	"encoding/gob"
)

type Creds struct {
	SenderId int
}

type RegisterReply struct {
	Id int
}

type RegisterArgs struct {
	Addrs []string
	Creds
}

type SetArgs struct {
	Name string
	Value interface{}
	Creds
}

type GetArgs struct {
	Name string
}

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

/* ********************* RPC Functions ********************* */
func (d *DSM) RegisterReplica(addr *string, reply *RegisterReply) error {
	// Registers a new replica with the DSM
	// Only called by the clients
	log.Printf("got register from %s\n", *addr)
	d.addrLock.Lock()
	defer d.addrLock.Unlock()
	d.addrs = append(d.addrs, *addr)
	reply.Id = len(d.addrs) - 1
	go func() {
		d.SendBroadcast("DSM.UpdateAddrs", &RegisterArgs{Addrs: d.addrs, Creds: Creds{SenderId: d.Id}})
	}()
	return nil
}

func (d *DSM) UpdateAddrs(args *RegisterArgs, reply *int) error {
	// Update the list of addresses
	// Only called by the registry service
	if args.SenderId == d.Id {
		return nil
	}
	log.Printf("Got update addrs: %v\n", args.Addrs)
	d.addrLock.Lock()
	defer d.addrLock.Unlock()
	d.addrs = args.Addrs
	*reply = 0
	return nil
}

func (d *DSM) GetVar(name string, reply *interface{}) error {
	// Get the value of a shared variable
	*reply = d.vars[name]
	return nil
}

func(d *DSM) SetVar(args *SetArgs, reply *int) error {
	// Set the value of a shared variable
	d.Set(args.Name, args.Value)
	*reply = 0
	return nil
}

func init() {
	gob.Register(RegisterArgs{})
	gob.Register(RegisterReply{})
	gob.Register(Creds{})
}

