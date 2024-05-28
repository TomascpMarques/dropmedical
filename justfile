prog_name := "dropmedical"
entry := "main"

# dev loop
dev:
  gow run {{entry}}.go

# Testa tudo, não renicia a db
t:
  go test ./... -v

# Testa um modulo singularmente
test MODULE: rebuild-db
  go test ./{{MODULE}}/... -v

# Corre o ficheiro com a função de entrada
run:
  go run {{entry}}.go

# Run the app with env variables
app:
  bash scripts/initApp.bash

# Constroi a aplicação e verifica para race conditions
build:
  go build --race -o target/{{prog_name}}

# Make sure you have run "set -a; source .env" !
migrate:
  migrate -database ${DATABASE_URL} -path ${MIGRATIONS} up

# Make sure you have run "set -a; source .env" !
createm NAME:
  migrate create -ext sql -dir ${MIGRATIONS} {{NAME}} ;\

# Reconstroi a base de dados
rebuild-db:
  docker stop sqlx-go; docker rm sqlx-go; just init-db;

# Docker build go api
dk-build:
  docker build -t dropmedical .

compose: dk-build
  docker compose up

# Incia a base de dados local
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
