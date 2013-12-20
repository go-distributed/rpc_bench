package rpc_bench

import (
	"log"
	"net"
	"net/http"
	"net/rpc"
	"strconv"
	"testing"
	"io"
)

const (
	ServerHTTP = 0
	ServerTCP  = 1
)

type Args struct {
	C []byte
}

type Foo int

func (t *Foo) Dummy(args *Args, reply *[]byte) error {
	*reply = args.C
	return nil
}

var start = 0

func BenchmarkHttpSync(b *testing.B) {
	done := make(chan bool, 10)
	if start != 1 {
		startRPC(ServerHTTP, &start)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10; i++ {
			go clientSync(ServerHTTP, start, done)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	}
}

func BenchmarkHttpAsync(b *testing.B) {
	done := make(chan bool, 10)
	if start != 2 {
		startRPC(ServerHTTP, &start)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10; i++ {
			go clientAsync(ServerHTTP, start, done)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	}
}

func BenchmarkTCPSync(b *testing.B) {
	done := make(chan bool, 10)
	if start != 3 {
		startRPC(ServerTCP, &start)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10; i++ {
			go clientSync(ServerTCP, start, done)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	}
}

func BenchmarkTCPAsync(b *testing.B) {
	done := make(chan bool, 10)
	if start != 4 {
		startRPC(ServerTCP, &start)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10; i++ {
			go clientAsync(ServerTCP, start, done)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	}
}

//func BenchmarkHttpDummySync(b *testing.B) {
//	done := make(chan bool, 10)
//	if start != 5{
//		startDummyRPC(ServerTCP, &start)
//	}
//
//	b.ResetTimer()
//	for i := 0; i < b.N; i++ {
//		for i := 0; i < 10; i++ {
//			go clientSync(ServerTCP, start, done)
//		}
//		for i := 0; i < 10; i++ {
//			<-done
//		}
//	}
//}

func BenchmarkTcpDummySync(b *testing.B) {
	done := make(chan bool, 10)
	if start != 5{
		startDummyRPC(ServerTCP, &start)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < 10; i++ {
			go clientDummy(ServerTCP, start, done)
		}
		for i := 0; i < 10; i++ {
			<-done
		}
	}
}

func startDummyRPC(stype int, start *int) {
	*start = *start + 1
	foo := new(Foo)
	port := strconv.Itoa(*start + 3000)

	rpc.Register(foo)
	/*if stype == ServerHTTP {
		http.DefaultServeMux = http.NewServeMux()
		rpc.Serve
	}*/

	if stype == ServerTCP {
		ln, err := net.Listen("tcp", "localhost:"+port)
		if err != nil {
			log.Fatal("listen error:", err)
		}
		go func() {
			for {
				conn, err := ln.Accept()
				if err != nil {
					log.Fatal("accept error:", err)
				}
				go func(conn io.ReadWriteCloser) { // ServeConn
					dum := &DummyCodec{conn, make([]byte, 4096)}
					rpc.ServeCodec(dum)
				}(conn)
			}
		}()
	}
}

func startRPC(stype int, start *int) {
	*start = *start + 1
	foo := new(Foo)
	port := strconv.Itoa(*start + 3000)

	rpc.Register(foo)
	if stype == ServerHTTP {
		http.DefaultServeMux = http.NewServeMux() // avoid panic for dup-registering handler
		rpc.HandleHTTP()
		go http.ListenAndServe("localhost:"+port, nil)
	}
	if stype == ServerTCP {
		ln, err := net.Listen("tcp", "localhost:"+port)
		if err != nil {
			log.Fatal("listen error:", err)
		}
		go rpc.Accept(ln)
	}
}

func clientDial(stype, pt int) *rpc.Client {
	port := strconv.Itoa(pt + 3000)
	if stype == ServerHTTP {
		client, err := rpc.DialHTTP("tcp", "localhost:"+port)
		if err != nil {
			log.Fatal("diaHTTP fail:", err)
		}
		return client
	}

	//TCP
	client, err := rpc.Dial("tcp", "localhost:"+port)
	if err != nil {
		log.Fatal("dialTCP fail:", err)
	}
	return client
}

func clientSync(stype, pt int, done chan bool) {
	content := make([]byte, 8192)
	for i := 0; i < len(content); i++ {
		content[i] = 32
	}
	
	client := clientDial(stype, pt)
	defer client.Close()

	args := &Args{content}
	var reply []byte

	for i := 0; i < 1000; i++ {
		err := client.Call("Foo.Dummy", args, &reply)
		if err != nil {
			log.Fatal("Dummy error:", err)
		}
	}
	//log.Println(reply)
	done <- true
}

func clientAsync(stype, pt int, done chan bool) {
	content := make([]byte, 8192)
	for i := 0; i < len(content); i++ {
		content[i] = 32
	}
	
	client := clientDial(stype, pt)
	defer client.Close()

	args := &Args{content}
	var reply []byte

	for i := 0; i < 1000; i++ {
		call := client.Go("Foo.Dummy", args, &reply, nil)
		<-call.Done
		if call.Error != nil {
			log.Fatal("Dummy error:", call.Error)
		}
		//log.Println(reply)
	}
	done <- true
}

func dialDummy(stype, pt int) *rpc.Client {
	port := strconv.Itoa(pt+3000)
	conn, err := net.Dial("tcp", "localhost:"+port)
	if err != nil {
		log.Fatal("dial err:", err)
	}
	return rpc.NewClientWithCodec(&DummyCodec{conn, make([]byte, 4096)})
}

func clientDummy(stype, pt int, done chan bool) {
	content := make([]byte, 8192)
	for i := 0; i < len(content); i++ {
		content[i] = 32
	}
	
	client := dialDummy(stype, pt)
	defer client.Close()

	args := &Args{content}
	var reply []byte

	for i := 0; i < 1000; i++ {
		err := client.Call("Foo.Dummy", args, &reply)
		if err != nil {
			log.Fatal("Dummy err:", err)
		}
	}
	done <- true
}