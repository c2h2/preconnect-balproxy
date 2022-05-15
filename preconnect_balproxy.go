package main

import (
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"sync/atomic"
	"time"
)

var spareMaxConnPool int
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
	_spareMaxConnPool := flag.Int("c", 64, "How many spare remote connections stay open in advance, default: 64")
	flag.Var(&remoteSrvs, "r", "remote address(es) use more than once, example: -r 127.0.0.1:3128 -r 127.0.0.1:8118")
	flag.Var(&remoteBackupSrvs, "rb", "remote backup address(es) use more than once, example: -rb 127.0.0.1:3128 -rb 127.0.0.1:8118") //TODO implement.

	flag.Parse()
	originalRemoteServs = remoteSrvs //copy to original servers
	spareMaxConnPool = *_spareMaxConnPool

	if len(remoteSrvs) == 0 {
		fmt.Println("Need at least one -r flag to run.\n example: go run balproxy.go -b 0.0.0.0:1234 -r 192.168.200.1:1077 -r 192.168.200.1:1078")
		os.Exit(1)
	}
	remts = make(chan net.Conn, spareMaxConnPool)
	sem = make(chan int, spareMaxConnPool)
	proxy, err := net.Listen("tcp", *bindAddr)
	fmt.Print("Listening " + *bindAddr + ", Remote Servers: ")
	fmt.Println(remoteSrvs)
	if err != nil {
		panic(err)
	}
	go initToRemote()
	go heartbeat()
	go healthChk()
	acceptFromProxy(proxy)
}

func heartbeat() {
	for {
		pconn := atomic.LoadInt32(activeRemoteTCPCnt)
		rconn := atomic.LoadInt32(activeProxyCnt)
		fmt.Printf("Proxy_Conn: %d, Remote_Conn: %d, Spare_Conn: %d, Up_Remotes: ", pconn, rconn, pconn-rconn)
		fmt.Println(remoteSrvs)
		time.Sleep(time.Second * 3)
	}
}

//healthChk all servers and temp ban failed servers
func healthChk() {
	for {
		var tempSrvs []string
		for i := 0; i < len(originalRemoteServs); i++ {
			for retry := MAXRETRY; retry > 0; retry-- {
				d := net.Dialer{Timeout: time.Second * 60}
				remt, err := d.Dial("tcp", originalRemoteServs[i])
				if err != nil { //failed
					time.Sleep(1 * time.Second)
					if remt != nil {
						remt.Close()
					}
					continue //next loop to retry
				} else {
					remt.Close()
					tempSrvs = append(tempSrvs, originalRemoteServs[i])
					break //working
				}
			}
			time.Sleep(time.Millisecond * 100) //next server
		}
		remoteSrvs = tempSrvs       //set to current servers
		time.Sleep(time.Second * 5) //next time to check
	}
}

func acceptFromProxy(proxy net.Listener) {
	totalConnected := 0
	for {
		activeCnt := atomic.AddInt32(activeProxyCnt, 1)
		remt := <-remts
		<-sem
		newlisten, err := proxy.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}
		go forwardToRemote(remt, newlisten)
		totalConnected++
		fmt.Printf("%d/%d: %v <-> %v <-> %v\n", activeCnt, totalConnected, newlisten.RemoteAddr(), newlisten.LocalAddr(), remt.RemoteAddr())
	}
}

//establish some connection to remote
func initToRemote() {
	for { //loop forever for creating new remote connections
		sem <- 1 //limit number of connections.
		go func() {
			_remoteSrvs := remoteSrvs
			upServers := len(_remoteSrvs)
			for { //loop until we get a good connection
				if upServers <= 0 {
					fmt.Println("initToRemote: no more servers.")
					time.Sleep(time.Second)
					break //no more servers
				}
				idx := rand.Intn(len(_remoteSrvs))
				remoteSrv := _remoteSrvs[idx]
				d := net.Dialer{Timeout: 60 * time.Second}
				remt, err := d.Dial("tcp", remoteSrv)
				if err != nil {
					fmt.Println(err)
					if remt != nil {
						remt.Close()
					}
					upServers--
					_remoteSrvs = remove(_remoteSrvs, idx)
					continue
				} else { //established
					atomic.AddInt32(activeRemoteTCPCnt, 1)
					remts <- remt
					break
				}
			}
		}()
	}
}

func forwardToRemote(remt net.Conn, newlisten net.Conn) {
	defer newlisten.Close()
	defer remt.Close()
	defer atomic.AddInt32(activeProxyCnt, -1)
	defer atomic.AddInt32(activeRemoteTCPCnt, -1)
	go io.Copy(remt, newlisten)
	io.Copy(newlisten, remt)
}
