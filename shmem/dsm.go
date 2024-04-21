package dsm

import (
	"log"
	"net/rpc"
	"sync"
	"time"
)

type UpdateMessage struct {
	Id        int64 // ID of the message
	Args      SetArgs
	Priority  int
	delivered bool
}

type PriorityQueue []*UpdateMessage

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *PriorityQueue) Push(x interface{}) {
	item := (x).(*UpdateMessage)
	log.Printf("Pushing item: %v, New length: %v", item, len(*pq))
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func (pq *PriorityQueue) Top() interface{} {
	old := *pq
	n := len(old)

	if n == 0 {
		return nil
	}
	
	log.Printf("Index: %v\n", n-1)
	item := old[n-1]
	return item
}

type DSM struct {
	Id          int      // Unique ID for the client
	addrs       []string // Addresses of all clients
	addrLock    sync.Mutex
	vars        map[string]*interface{} // Shared variables
	maxPriority int
	pq          PriorityQueue
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

		for len(d.addrs) == 0 {
			time.Sleep(1 * time.Second)
		}
	}
	return nil
}

func (d *DSM) getim(name string) *interface{} {
	// Get the pointer to a shared variable
	// This returns nil if the variable does not exist
	return d.vars[name]
}

func (d *DSM) setim(name string, value interface{}) {
	// Set the value of a shared variable
	if d.vars[name] == nil {
		d.vars[name] = &value
		} else {
			*d.vars[name] = value
		}
}

func (d *DSM) Get(name string) *interface{} {
		// Get the value of a shared variable
		return d.getim(name)
}

func (d *DSM) Set(name string, value interface{}) {
	// Send a write update to all clients
	messageId := time.Now().UnixNano()
	updateMsg := UpdateMessage{Id: messageId, Args: SetArgs{Name: name, Value: value, Creds: Creds{SenderId: d.Id}}, delivered: false}
	log.Printf("[DEBUG] Sending write update: %v\n", updateMsg)
	replies := d.SendBroadcast("DSM.ProposePriority", updateMsg)

	// Find the max priority
	maxPriority := d.maxPriority
	for _, p := range replies {
		if p > maxPriority {
			maxPriority = p
		}
	}

	d.maxPriority = maxPriority
	
	log.Printf("[DEBUG] Got max priority: %v\n", maxPriority)

	// Send the final priority
	updateMsg.Priority = maxPriority
	d.SendBroadcast("DSM.FinalPriority", updateMsg)
	log.Printf("[DEBUG] Sent final priority %v \n", updateMsg.Priority)
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
		log.Printf("[DEBUG] Calling %s(%s) on %s\n", fun, args, addr)
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
