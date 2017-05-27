package main

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Declare a global variable to store the Redis connection pool.
var POOL *redis.Pool
var PORT = "4000"
var SIGN_KEY = []byte("secret")

//Init just loads the default variables
func init() {
	POOL = newPool("localhost:6379")
	PORT = os.Getenv("AUTH_PORT")
	SIGN_KEY = []byte(os.Getenv("SIGN_KEY"))
}

// returns a pointer to a redis pool
func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

// gets you a token if you pass the right credentials
func login(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Method != "POST" {
		http.Error(w, "Forbidden", 403)
		return
	}
	// Authenticate
	conn := POOL.Get()
	defer conn.Close()
	email := strings.ToLower(r.FormValue("email"))
	password, err := redis.Bytes(conn.Do("GET", email))
	if err != nil {
		log.Println(err)
		return
	}
	// if password doesn't match
	if string(password[:]) != r.FormValue("password") {
		http.Error(w, "Forbidden", 403)
		return
	}
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["admin"] = false
	// don't forget to validate this string
	claims["email"] = email
	// 24 hour token
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	tokenString, _ := token.SignedString(SIGN_KEY)
	w.Write([]byte(tokenString))
}

// register a new user, gives you a token if the email -> password
// is not registered already
func register(w http.ResponseWriter, r *http.Request) {
	var err error
	if r.Method != "POST" {
		http.Error(w, "Forbidden", 403)
		return
	}
	conn := POOL.Get()
	defer conn.Close()
	email := strings.ToLower(r.FormValue("email"))
	// check if the user is already registered
	exists, err := redis.Bool(conn.Do("EXISTS", email))
	if exists {
		w.Write([]byte("Email taken"))
		return
	}
	password := r.FormValue("password")
	// Set user -> password in redis
	_, err = conn.Do("SET", email, string(password[:]))
	if err != nil {
		log.Println(err)
	}
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["admin"] = false
	// don't forget to validate this string
	claims["email"] = email
	// 24 hour token
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	tokenString, _ := token.SignedString(SIGN_KEY)
	w.Write([]byte(tokenString))
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/login", login)
	r.HandleFunc("/register", register)
	srv := &http.Server{
		Handler:      r,
		Addr:         ":" + PORT,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
