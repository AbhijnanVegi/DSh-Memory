package dsm

import (
	"log"
	"sync"
	"net/rpc"
)

type RegisterReply struct {
	Addrs []string
}

type BroadcastArgs struct {
	Ptr int
}

type DSM struct {
	addrs []string
	addrLock sync.Mutex
}

func (d *DSM) Init(serverAddress string, myAddress string) error {
	if serverAddress == "" {
		return nil
	} else {
		client, err := rpc.DialHTTP("tcp", serverAddress)
		if err != nil {
			return err
		}
		var reply RegisterReply
		err = client.Call("DSM.RegisterReplica", &myAddress, &reply)
		if err != nil {
			return err
		}
		d.addrLock.Lock()
		defer d.addrLock.Unlock()
		d.addrs = reply.Addrs
		d.addrs = append(d.addrs, serverAddress)
		log.Printf("Got addrs: %v\n", d.addrs)
	}
	return nil
}

func (d *DSM) RegisterReplica(addr *string, reply *RegisterReply) error {
	log.Printf("got register from %s\n", *addr)
	d.addrs = append(d.addrs, *addr)
	reply.Addrs = d.addrs
	return nil
}

func (d *DSM) Update(args *BroadcastArgs, reply *int) error {
	log.Printf("got update %s\n", args.Ptr)
	*reply = 0
	return nil
}

func (d *DSM) SendBroadcast() {
	log.Printf("[DEBUG] Sending broadcast to %s\n", d.addrs)
	for _, addr := range d.addrs {
		client, err := rpc.DialHTTP("tcp", addr)
		if err != nil {
			log.Printf("Error dialing %s: %v\n", addr, err)
			continue
		}
		var reply int
		args := BroadcastArgs{Ptr: 5}
		log.Printf("[DEBUG] Sending broadcast to %s\n", addr)
		err = client.Call("DSM.Update", &args, &reply)
		if err != nil {
			log.Printf("Error calling %s: %v\n", addr, err)
			continue
		}
	}
}

