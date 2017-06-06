# Minimalistic Go JWT Auth example

#### Dependencies
+ `github.com/dgrijalva/jwt-go`
+ `github.com/garyburd/redigo/redis`
+ `github.com/gorilla/mux`
+ `golang.org/x/crypto/bcrypt`

#### Installation
+ Install [redis](https://redis.io)
+ Clone repo to `$GOPATH/src/github.com/octohedron/gauth`
+ Install dependencies
```Bash
$ go get
```
+ Set environment variables
```Bash
$ export AUTH_PORT=YOUR_PORT # i.e. 8000
$ export SIGN_KEY=secret # your jwt sign key
```
#### Usage
```Bash
$ go build && ./gauth
```
This will run the server and you can try it out with curl

#### Register
```Bash
$ curl -X POST -F 'email=a@a.com' -F 'password=password' http://192.168.1.43:4200/register
# Should print out a token, similar to 
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.ey...
```


#### Login
```Bash
$ curl -X POST -F 'email=a@a.com' -F 'password=password' http://192.168.1.43:4200/login
# Should print out a token, similar to 
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.ey...
```

LICENSE: MIT