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
The implemented algorithm for sharing memory ensures sequential consistency, i.e, the result of any execution of the operations of all the processors is the same as if they were executed in a sequential order.

## Coherence protocols
The implementation uses the Write-Update protocol for coherence. That is all the writes are broadcasted to the processors.

## Replication
This implementation replicates all the variables across all the processors, thanks to the Write-Update protocol. This allows multiple reads and writes with little overhead.