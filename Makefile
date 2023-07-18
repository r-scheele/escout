CONTAINER_NAME=escout-database
DB_URL=postgresql://root:secret@192.168.165.2:5432/escout?sslmode=disable


network:
	docker network create escout-network
connect-database:
	docker network connect escout-default "$(CONTAINER_NAME)"
connect-api:
	docker network connect escout-default escout-api-1

postgres:
	docker run --name "$(CONTAINER_NAME)" --network escout_default -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:14-alpine

mysql:
	docker run --name mysql8 -p 3306:3306  -e MYSQL_ROOT_PASSWORD=secret -d mysql:8

createdb:
	docker exec -it "$(CONTAINER_NAME)" createdb --username=root --owner=root escout

dropdb:
	docker exec -it "$(CONTAINER_NAME)" dropdb escout

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

sqlc:
	sqlc generate

test:
	go test -v -cover -short ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/r-scheele/escout/db/sqlc Store
	mockgen -package mockwk -destination worker/mock/distributor.go github.com/r-scheele/escout/worker TaskDistributor


evans:
	evans --host localhost --port 9090 -r repl

redis:
	docker run --name redis -p 6379:6379 -d redis:7-alpine

.PHONY: network connect-network postgres createdb dropdb migrateup migratedown migrateup1 migratedown1 new_migration db_docs db_schema sqlc test server mock proto evans redis
