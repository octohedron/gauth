# Minimalistic Go JWT Authentication example

#### Dependencies
+ `github.com/dgrijalva/jwt-go`
+ `github.com/garyburd/redigo/redis`
+ `github.com/gorilla/mux`

#### Installation
+ Install [redis](https://redis.io)
+ Clone repo to `$GOPATH/src/github.com/octohedron/gauth`
+ Instlal dependencies
```Bash
$ go get
```
+ Build the binary
```Bash
$ go build
```
+ Set environment variables
```Bash
export AUTH_PORT=YOUR_PORT # i.e. 80
export SIGN_KEY=secret # your jwt sign key
```
#### Usage
```Bash
go build && ./gauth
```
This will run the server and you can try it out with curl

#### Register
```Bash
curl -X POST -F 'email=a@a.com' -F 'password=password' http://192.168.1.43:4200/register
# Should print out a token, similar to 
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6ZmFsc2UsImVtYWlsIjoiYUB0LmNvbSIsImV4cCI6MTQ5NjA4OTQ1M30.qWuLX8pYA1RF83ogvMivs5yDqqx5szlvj_eG2jp-H2kM
```


#### Login
```Bash
curl -X POST -F 'email=a@a.com' -F 'password=password' http://192.168.1.43:4200/login
```

LICENSE: MIT