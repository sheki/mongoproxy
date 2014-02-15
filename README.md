mongoproxy
==========

A proxy to better connection pool to MongoDB.

#Build
* Checkout the code.
* Set the GOPATH to the checkout ROOT.

<code>
go get ./...
</code>

<code>
go install ./mongoproxybin
</code>

#Test

<code>
go get -t ./...
</code>

<code>
go test ./...
</code>

#Running the binary
The $GOPATH/bin directory contains the binary
<code>
mongoproxybin 
</code>
## Config/Flags 
* -dispatch_queue_len=10000: dispatch queue length
* -listen_addr=":6000": address to listen for incomming requests
* max_mongo_conn=100: max connections to mongo
* -mongo_addr=":27017": address of the mongo to proxy
* -read_timeout=10s: read time out for connections
* -write_timeout=10s: write time out for connections
