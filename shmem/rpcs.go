package dsm

import (
	"container/heap"
	"errors"
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

func (d *DSM) SetVar(args *SetArgs, reply *int) error {
	// Set the value of a shared variable
	d.Set(args.Name, args.Value)
	*reply = 0
	return nil
}

func (d *DSM) ProposePriority(message UpdateMessage, reply *int) error {
	// Propose a priority for a message
	d.maxPriority = d.maxPriority + 1
	log.Printf("Proposing priority %v for message %v\n", d.maxPriority, message)
	*reply = d.maxPriority
	heap.Push(&d.pq, &UpdateMessage{priority: d.maxPriority, Id: message.Id, Args: message.Args, delivered: false})
	return nil
}

func (d *DSM) FinalPriority(message UpdateMessage, reply *int) error {

	log.Printf("Finalizing priority for message %v\n", message)

	*reply = 0
	messageFound := false
	// Update the priority of a message
	for i, v := range d.pq {
		if v.Id == message.Id {
			d.pq[i].priority = message.priority
			d.pq[i].delivered = true
			heap.Fix(&d.pq, i)
			messageFound = true
			break
		}
	}

	if !messageFound {
		return errors.New("message not found")
	}

	highestPriorityMessage := d.pq.Top().(*UpdateMessage)
	if highestPriorityMessage == nil {
		return nil
	}
	for highestPriorityMessage.delivered {
		log.Printf("Delivering message: %v\n", highestPriorityMessage)
		d.Set(highestPriorityMessage.Args.Name, highestPriorityMessage.Args.Value)
		d.pq.Pop()
		_highestPriorityMessage := d.pq.Top()

		if _highestPriorityMessage == nil {
			break
		} else {
			highestPriorityMessage = _highestPriorityMessage.(*UpdateMessage)
		}

		*reply = *reply + 1
	}

	return nil
}
