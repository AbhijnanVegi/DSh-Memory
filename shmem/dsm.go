package dsm

import (
	"log"
	"sync"
	"net/rpc"
)

type RegisterReply struct {
	Addrs []string
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
