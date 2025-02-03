gen:
	protoc -I=proto proto/*.proto --go_out=. --go-grpc_out=.
clean:
	rm pb/*.go
test:
	go test -cover -race ./...
server:
	go run cmd/server/main.go -port 8080
client:
	go run cmd/client/main.go -address 0.0.0.0:8080