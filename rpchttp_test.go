package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"testing"
)

type Args struct {
	C string
}

type Foo int

func (t *Foo) Dummy(args *Args, reply *string) error {
	*reply = "hello " + args.C
	return nil
}

var start = false

func BenchmarkHttpSync(b *testing.B) {
	done := make(chan bool, 10)
	fmt.Println("start")
	if !start {
		startHttpRPC()
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10; i++ {
			go clientSync(done)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	}
}

func startHttpRPC() {
	start = true
	foo := new(Foo)

	rpc.Register(foo)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

func clientHttpDial() *rpc.Client {
	client, err := rpc.DialHTTP("tcp", "localhost:1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	return client
}

func clientSync(done chan bool) {
	client := clientHttpDial()
	defer client.Close()

	args := &Args{"yifan"}
	var reply string

	err := client.Call("Foo.Dummy", args, &reply)
	if err != nil {
		log.Fatal("Dummy error:", err)
	}
	fmt.Println(reply)
	done <- true
}

func clientAsync(done chan bool) {
	client := clientHttpDial()
	defer client.Close()

	args := &Args{"yifan"}
	var reply string

	call := client.Go("Foo.Dummy", args, &reply, nil)
	<-call.Done
	if call.Error != nil {
		log.Fatal("Dummy error:", call.Error)
	}
	fmt.Println(reply)
	done <- true
}
