package dsm

import (
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

func init() {
	gob.Register(RegisterArgs{})
	gob.Register(RegisterReply{})
	gob.Register(Creds{})
}