mongoproxy
==========

MongoDB has a limit of 20,000 connections per node. When you have a lot of App-server talking to Mongo you will hit this limit. This problem gets worse in Forked-enviroments like Ruby's Unicorn where every app node opens multiple non sharable connections. To solve this we run a mongoproxy in the middle which acts converts m incoming connections to n outgoing connections to mongo keeping (m > n).

#Known Issues
* Mongo's [GetLastError](http://docs.mongodb.org/manual/reference/command/getLastError/) is local to the connection, the requests are made on, the proxy currently does not have any connection affinity and this breaks stuff.
* Mongo ReplicaSet commands will return the original hosts and not the proxies in front of the host, this will cause the clients to talk to the actual mongo nodes and not proxies.

#Build
```sh
$ go get github.com/sheki/mongoproxy/cmd/mongoproxy
```

#Test

```sh
$ go get -t github.com/sheki/mongoproxy
$ go test github.com/sheki/mongoproxy
```

#Running the binary
go get will place a 'mongoproxy' binary in the bin subdirectory of the first path in $GOPATH

## Config/Flags 
* -dispatch_queue_len=10000: dispatch queue length
* -listen_addr=":6000": address to listen for incomming requests
* max_mongo_conn=100: max connections to mongo
* -mongo_addr=":27017": address of the mongo to proxy
* -read_timeout=10s: read time out for connections
* -write_timeout=10s: write time out for connections
