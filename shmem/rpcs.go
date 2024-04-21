package dsm

import (
	"log"
)

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
	*reply = d.Get(name)
	return nil
}

func(d *DSM) SetVar(args *SetArgs, reply *int) error {
	// Set the value of a shared variable
	d.Set(args.Name, args.Value)
	*reply = 0
	return nil
}