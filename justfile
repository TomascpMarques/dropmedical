prog_name := "dropmedical"
entry := "main"

dev:
  gow run {{entry}}.go

test: rebuild-db init-db 
  go test ./... -v

t:
  go test ./... -v

run:
  go run {{entry}}.go

build:
  go build --race -o target/{{prog_name}}

# Make sure you have run "set -a; source .env" !
migrate: 
  migrate -database ${DATABASE_URL} -path ${MIGRATIONS} up

# Make sure you have run "set -a; source .env" !
createm NAME: 
  migrate create -ext sql -dir ${MIGRATIONS} {{NAME}} ;\

rebuild-db:
  docker stop sqlx-go; docker rm sqlx-go;

init-db:
  docker run \
        --name sqlx-go \
        -e POSTGRES_HOST="localhost" \
        -e POSTGRES_USER="postgres" \
        -e POSTGRES_PASSWORD="password" \
        -e POSTGRES_DB="main" \
        -p "5555":5432 \
        -d postgres \
        postgres -N 1000;\
  sleep 1.1;
