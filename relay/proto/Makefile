protoc:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative relay.proto

clean:
	rm -rf relay.pb.go
	rm -rf relay_grpc.pb.go
