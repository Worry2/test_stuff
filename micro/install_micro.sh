go get -u github.com/micro/micro
go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
go get -u github.com/micro/protoc-gen-micro

# Set $GOPATH to be able to run micro new

# Install protobuf: get source and compile:
# https://github.com/google/protobuf
# ./configure
# ./autogen.sh
# make
# make install

# NOTES: mdns registry does not work well...
# Install consul
# download, unzip
consul agent -dev
# or docker run consul

# MICRO_API_HANDLER=rpc MICRO_API_NAMESPACE=go.micro.srv micro api
# curl -XPOST -H 'Content-Type: application/json' -d '{"name": "John"}' http://localhost:8080/example/call