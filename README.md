# Distributed-Shared-Memory
A simple distributed shared memory system implemented in Go.

# Running the program
1. Compile the program using `go build`
2. Run the registry server `./dsm :6000 :6001`
3. Run the clients with 3rd argument being the adress of the registry server's rpc endpoint
`./dsm :7000 :7001 localhost:6001` `./dsm :8000 :8001 localhost:6001`
4. Create endpoints to test shared memory

# Source code
`./shmem` contains the library code for distributed shared memory implementation. `./main.go` is sample web server written using this library to demo the usage of this library.

# Design
## Client Discovery
A registry is setup which is then connected to by the clients joining the network. This registry service acts to inform the clients of the other clients in the network and also updating them incase clients join/leave.

## Variable sharing
Since Go, doesn't provide low level memory access, we implement shared memory through sharing variables. The DSM object consists of two methods `Get` and `Set` to get and set the values of shared variables respectively

## Consistency
We ensure Sequential Consistency by utilizing the Total Order Broadcast. This is accomplished by using the ISIS algorithm for Total Order Broadcast. The ISIS algorithm is a distributed algorithm that ensures that all the messages are delivered in the same order to all the processes. 

## Write Update 
We follow the Full-Replication Distributed Shared Memory architecture, which involves sending a write update message to all the clients involved in the replication. To ensure consistency, we use the Total Order Multicast protocol to ensure that all the clients receive the write update message in the same order. This ensures Sequential Consistency. 

## ISIS Algorithm
The ISIS algorithm involves sending the update message twice. 
1. The first message is sent to all the processes and each process replies with a proposed priority and marks the message the messaged as undelivered. 
2. Each sender ensures that the proposed priority is greater than the previously proposed priorities and the previously agreed priorities.
> Two senders can decide on the same priority, in which case the sender with the lower id is given the priority. (Sender id is used as a tie breaker)
3. The sender then decides on the priority of the message (maximum) and sends the message again with the priority. 
4. Each receiver modificies the priority of the existin message and then mark it as delivered.
5. Then the receiver delivers any delivered messages at the front of the queue.


## Coherence protocols
The implementation uses the Write-Update protocol for coherence. That is all the writes are broadcasted to the processors.

## Replication
This implementation replicates all the variables across all the processors, thanks to the Write-Update protocol. This allows multiple reads and writes with little overhead.