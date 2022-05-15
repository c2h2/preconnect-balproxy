package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
)

var maxConnPool int
var remts chan net.Conn
var sem chan int
var activeProxyCnt *int32 = new(int32)
var activeRemoteTCPCnt *int32 = new(int32)
var MAXRETRY = 3

type arrayFlags []string

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func remove(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

var originalRemoteServs []string
var remoteSrvs arrayFlags
var remoteBackupSrvs arrayFlags

func main() {
	bindAddr := flag.String("b", "0.0.0.0:1234", "Program bind address, default 0.0.0.0:1234")
	maxConnPool = *flag.Int("c", 64, "How many remote connections stay open in advance, default: 64")
	flag.Var(&remoteSrvs, "r", "remote address(es) use more than once, example: -r 127.0.0.1:3128 -r 127.0.0.1:8118")
	//flag.Var(&remoteBackupSrvs, "rb", "remote backup address(es) use more than once, example: -rb 127.0.0.1:3128 -rb 127.0.0.1:8118") //TODO implement.

	flag.Parse()
	originalRemoteServs = remoteSrvs //copy to original servers

	if len(remoteSrvs) == 0 {
		fmt.Println("Need at least one -r flag to run.\n example: go run basic_balproxy.go -b 0.0.0.0:1234 -r 192.168.200.1:1077 -r 192.168.200.1:1078")
		os.Exit(1)
	}

	ln, err := net.Listen("tcp", *bindAddr)
	fmt.Print("Listening ", *bindAddr, ", Remote Servers: ", remoteSrvs)

	if err != nil {
		panic(err)
	}

	i := 0
	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}
		i++
		go handleRequest(i, conn, remoteSrvs)
	}
}

func handleRequest(i int, conn net.Conn, remoteSrvs []string) {
	defer conn.Close()
	idx := rand.Intn(len(remoteSrvs))
	remoteSrv := remoteSrvs[idx]
	proxy, err := net.Dial("tcp", remoteSrv)
	if err != nil {
		fmt.Println(err)
		handleRequest(i, conn, remoteSrvs)
		return
	}

	defer proxy.Close()
	go io.Copy(proxy, conn)
	fmt.Printf("%d: %v <-> %v <-> %v\n", i, conn.LocalAddr(), conn.RemoteAddr(), remoteSrv)
	io.Copy(conn, proxy)
}
