package dsm

import (
	"log"
	"sync"
	"net/rpc"
	"encoding/gob"
)

type RegisterReply struct {
	Id int
}

type Creds struct {
	SenderId int
}

type RegisterArgs struct {
	Addrs []string
	Creds
}

type DSM struct {
	id int
	addrs []string
	addrLock sync.Mutex
}

func (d *DSM) Init(serverAddress string, myAddress string) error {
	if serverAddress == "" {
		d.id = 0
		d.addrs = []string{myAddress}
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
		d.id = reply.Id
		log.Printf("Got addrs: %v\n", d.addrs)
	}
	return nil
}

func (d *DSM) RegisterReplica(addr *string, reply *RegisterReply) error {
	log.Printf("got register from %s\n", *addr)
	d.addrLock.Lock()
	defer d.addrLock.Unlock()
	d.addrs = append(d.addrs, *addr)
	reply.Id = len(d.addrs) - 1
	go func() {
		d.SendBroadcast("DSM.UpdateAddrs", &RegisterArgs{Addrs: d.addrs, Creds: Creds{SenderId: d.id}})
	}()
	return nil
}

func (d *DSM) UpdateAddrs(args *RegisterArgs, reply *int) error {
	if args.SenderId == d.id {
		return nil
	}
	log.Printf("Got update addrs: %v\n", args.Addrs)
	d.addrLock.Lock()
	defer d.addrLock.Unlock()
	d.addrs = args.Addrs
	*reply = 0
	return nil
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

func init() {
	gob.Register(RegisterArgs{})
	gob.Register(RegisterReply{})
	gob.Register(Creds{})
}

