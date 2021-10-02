nginx/server.crt:
	openssl genrsa -out nginx/server.key 2048
	openssl req -new -key nginx/server.key -out nginx/server.csr -subj "/C=XX/ST=XX/L=XX/O=XX/OU=XX/CN=nginx"
	openssl x509 -days 3650 -req -signkey nginx/server.key -in nginx/server.csr -out nginx/server.crt

build: nginx/server.crt
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o example .

local: build
	SERVER_URL=http://0.0.0.0:8080/v1 SERVER_ADDR=0.0.0.0:8080 ./example

local-test:
	curl -s http://0.0.0.0:8080/api-docs | jq .servers[0].url
	curl --include http://0.0.0.0:8080/v1/pets

docker: build
	docker-compose up

docker-down:
	docker-compose down

docker-test:
	curl -k -s https://0.0.0.0:8080/api-docs | jq .servers[0].url
	curl -k --include https://0.0.0.0:8080/v1/pets
