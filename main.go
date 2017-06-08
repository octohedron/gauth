package main

import (
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Declare a global variable to store the Redis connection pool.
var POOL *redis.Pool
var PORT = ""
var SIGN_KEY = []byte("secret")

// Loads the default variables
func init() {
	POOL = newPool("localhost:6379")
	PORT = os.Getenv("AUTH_PORT")
	if PORT == "" {
		PORT = "4000"
	}
	SIGN_KEY = []byte(os.Getenv("SIGN_KEY"))
}

// Returns a pointer to a redis pool
func newPool(addr string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 240 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", addr) },
	}
}

// Gets you a token if you pass the right credentials
func login(w http.ResponseWriter, r *http.Request) {
	var err error
	setHeaders(w)
	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", "Forbidden request"), 403)
		return
	}
	conn := POOL.Get()
	defer conn.Close()
	email := strings.ToLower(r.FormValue("email"))
	password, err := redis.Bytes(conn.Do("GET", email))
	if err == nil {
		// compare passwords
		err = bcrypt.CompareHashAndPassword(password, []byte(r.FormValue("password")))
		// if it doesn't match
		if err != nil {
			http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", "Wrong password"), 401)
			return
		}
		token := jwt.New(jwt.SigningMethodHS256)
		claims := token.Claims.(jwt.MapClaims)
		claims["admin"] = false
		claims["email"] = email
		// 24 hour token
		claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
		tokenString, _ := token.SignedString(SIGN_KEY)
		w.Write([]byte(fmt.Sprintf("{ \"access_token\": \"%s\" }", tokenString)))
	} else {
		// email not found
		http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", "Email not found"), 401)
		return
	}
}

// Register a new user, gives you a token and sets the email -> password
// in redis if the email doesn't exist
func register(w http.ResponseWriter, r *http.Request) {
	var err error
	setHeaders(w)
	if r.Method != "POST" {
		http.Error(w, fmt.Sprintf("{ \"error\": \"%s\" }", "Forbidden request"), 403)
		return
	}
	conn := POOL.Get()
	defer conn.Close()
	email := strings.ToLower(r.FormValue("email"))
	// check if the user is already registered
	exists, err := redis.Bool(conn.Do("EXISTS", email))
	if exists {
		w.Write([]byte(fmt.Sprintf("{ \"error\": \"%s\" }", "Email taken")))
		return
	}
	// get password from the post request form value
	password := r.FormValue("password")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password),
		bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	// Set email -> password in redis
	_, err = conn.Do("SET", email, string(hashedPassword[:]))
	if err != nil {
		log.Println(err)
	}
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["admin"] = false
	claims["email"] = email
	// 24 hour token
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()
	tokenString, _ := token.SignedString(SIGN_KEY)
	w.Write([]byte(fmt.Sprintf("{ \"access_token\": \"%s\" }", tokenString)))
}

func setHeaders(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Content-Type", "application/json")
	return w
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
