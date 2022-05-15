# preconnect-balproxy
Go TCP load balancing with preconnect for lower latency. it will preconnect to remote server(s). and detect dead connections.

This works like HAProxy, but with maintaining a pool of pre-exisiting connections to remote server(s). Latency of any new local connections will be reduced by around a PING cost (50ms PING -> around 100ms reduction). some features:

* pre-establish TCP connection, reduce TCP handshake time.
* minimal cpu and memory impact.
* utilize multicore CPUs.
* handling very high amount of concurrency. (10K+)
* forwarding data with kernel level effiency.

## Use cases
* backend load balancing
* remote hosts (high ping, possible reset)
* combine with SSR, v2ray, trojan-go, navieproxy to enhance high availability, lower latency. 
* works on MAC, linux, openWRT

## Build
```./build_all_platforms.sh``` 
or download from [Releases](https://github.com/c2h2/preconnect-balproxy/releases).

## Run
listen ipv4 :1234, randomly forward to one of three servers, with 50 spare preconnected connections.

```builds/preconnect_balproxy-linux-arm64 -b 0.0.0.0:1234 -r 127.0.0.1:1077 -r 192.168.200.1:1077 -r 192.168.200.1:10123 -c 50```

## Tune on linux/openwrt
You may want to tune up the limit of open files
```ulimit -n 99999``` 

## TODO
* to handle subscribe link and automatically combine links with the same encryptions.
