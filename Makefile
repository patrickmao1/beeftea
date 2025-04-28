.PHONY: proto
proto:
	protoc -I proto/ \
		--go_out=./ \
		--go-grpc_out=require_unimplemented_servers=false:./ \
		proto/*.proto

up:
	docker compose up --build -d
	docker compose logs -f

down:
	docker compose down