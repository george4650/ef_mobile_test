
run: 
	go run cmd/rest/main.go

runGrpcServer: 
	go run cmd/grpc/main.go

runGrpcClient: 
	go run example_grpc/main.go

swag: swag_fmt
	swag init -d ./cmd/rest,./ 

swag_fmt:
	swag fmt