.PHONY: up down migrate run generate sqlc docker protogen wire server
generate:
	go tool gqlgen generate cmd/ordersystem/wire_gen.go
server:
	go run cmd/ordersystem/main.go
up:	
	migrate -path=sql/migrations -database "mysql://root:root@tcp(localhost:3306)/orders" -verbose up
down:
	migrate -path=sql/migrations -database "mysql://root:root@tcp(localhost:3306)/orders" -verbose down
migrate:
	migrate create -ext=sql -dir sql/migrations -seq $(name)
sqlc:
	sqlc generate
docker:
	docker-compose up -d
protogen:
	protoc --go_out=. --go-grpc_out=. internal/infra/grpc/protofiles/*.proto
evans:
	evans --proto internal/infra/grpc/protofiles/*.proto repl
wire:
	wire ./cmd/ordersystem/wire.go
run:
	go run cmd/ordersystem/main.go cmd/ordersystem/wire_gen.go
