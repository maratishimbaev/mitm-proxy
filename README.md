# MITM Proxy

## Installation
```
git clone https://github.com/maratishimbaev/mitm-proxy.git
```

## Usage
Generate certs:
```
./cert.sh
```

Run application:
```
docker-compose up
go run cmd/main.go
```

### Proxy
Proxy is available at :8000 port

### Repeater
Repeater is available at :8001 port

Method | Path | Description
------ | ---- | -----------
GET | /requests | Get all requests
GET | /requests/{id:[0-9]+} | Repeat request with id
GET | /requests/{id:[0-9]+}/check | Check XXE vulnerability