## Overview


Open-API Server rest path:

|path|info|
|-|-|
|/api-docs|spec doc
|/v1/pets|list pets

## How to use.

Execute the following command in the root directory of git.

#### Local Server Test:

1. Start local server:

    make local

2. Connect to local server:

    make local-test

Succeeds as follows.

```
curl -s http://0.0.0.0:8080/api-docs | jq .servers[0].url
"http://0.0.0.0:8080/v1"
curl --include http://0.0.0.0:8080/v1/pets
HTTP/1.1 200 OK
Date: Sat, 02 Oct 2021 13:45:57 GMT
Content-Length: 0
```

#### Reverse Proxied Server Test:

Realized in docker environment for testing

- front-server: https://0.0.0.0:8080/ (nginx)
- openapi-server http://goserver:8080/

/etc/nginx/conf.d/default.conf:
```
...
location / {
    proxy_pass    http://goserver:8080/;
}
...
```

1. Start docker server:

    make docker

2. Connect to docker server:

    make docker-test

Fails as follows.

```
curl -k -s https://0.0.0.0:8080/api-docs | jq .servers[0].url
"https://0.0.0.0:8080/v1"
curl -k --include https://0.0.0.0:8080/v1/pets
HTTP/1.1 500 Internal Server Error
Server: nginx/1.21.3
Date: Sat, 02 Oct 2021 13:37:32 GMT
Content-Length: 0
Connection: keep-alive
```

Server console log:

```
goserver_1  | 2021/10/02 13:37:32 routeRegexp:path /v1/pets ^/v1/pets$
goserver_1  | 2021/10/02 13:37:32 methodMatcher
goserver_1  | 2021/10/02 13:37:32 schemeMatcher
goserver_1  | 2021/10/02 13:37:32 1 
goserver_1  | 2021/10/02 13:37:32 2 [https] http false
goserver_1  | 2021/10/02 13:37:32 DONT MATCH [https] &{<nil> <nil> map[] <nil>}
goserver_1  | 2021/10/02 13:37:32 routeRegexp:path /v1/pets ^/v1/pets/(?P<v0>[^/]+)$
goserver_1  | 2021/10/02 13:37:32 DONT MATCH &{/v1/pets/{petId} 0 {false true} 0xc000101b80 /v1/pets/%s [petId] [0xc000101a40] false} &{<nil> <nil> map[] <nil>}
```