swagger:
	swag init -g cmd/api/main.go -o docs

migrate-up:
	migrate -path migrations \
    -database "postgres://postgres:postgres@localhost:5432/myapp?sslmode=disable" \
    up

migrate-down:
	migrate -path migrations \
    -database "postgres://postgres:postgres@localhost:5432/myapp?sslmode=disable" \
    down

migrate-fresh:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/myapp?sslmode=disable" drop -f
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/myapp?sslmode=disable" up

migration:
	migrate create -ext sql -dir migrations -seq $(name)

migrate-reset:
	migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/myapp?sslmode=disable" drop -f
