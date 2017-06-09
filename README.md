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
{ "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ... " }
```


#### Login
```Bash
$ curl -X POST -F 'email=a@a.com' -F 'password=password' http://192.168.1.43:4200/login
# Should print out a token, similar to 
{ "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ... " }
```

Example AJAX usage

```JavaScript
//
// Note: For this to work you might need to uncomment the CORS headers in the
// setHeaders func
//
// button, on click ... 
var formData = new FormData();
formData.append("email", "a@a.com");
formData.append("password", "hunter2");
// make the request
fetch("http://192.168.1.43:4200/login", {
  method: "POST",
  body: formData
}).then(result => {
  result.json().then(result => {
    if (result.error) {
      alert(result.error); // "Wrong password" or "Email not found"
    } else {
      alert("Your token is: " + result.access_token);
    }
  });
});
```

LICENSE: MIT