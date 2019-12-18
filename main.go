package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

// server type is for sharing dependencies with handlers
type server struct {
	pool   *redis.Pool
	router *mux.Router
}

// POOL - Declare a global variable to store the Redis connection pool.
var POOL *redis.Pool

// PORT - The running port for this service
var PORT = ""

// SIGN_KEY - The JWT Sign key
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

func (s *server) routes() {
	s.router.HandleFunc("/login", s.handleLogin())
	s.router.HandleFunc("/register", s.handleRegister())
}

// Gets you a token if you pass the right credentials
func (s *server) handleLogin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
}

// Register a new user, gives you a token and sets the email -> password
// in redis if the email doesn't exist
func (s *server) handleRegister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
}

func setHeaders(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Set("Content-Type", "application/json")
	// Uncomment the following lines if you are having CORS issues,
	// optionally, replace "*" with your preferred address
	//
	// w.Header().Set("Access-Control-Allow-Origin", "*")
	// w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	// w.Header().Set("Access-Control-Allow-Headers",
	// 	"Accept, 0, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	return w
}

func main() {
	sr := &server{
		router: mux.NewRouter(),
		pool:   POOL,
	}
	sr.routes()
	srv := &http.Server{
		Handler:      sr.router,
		Addr:         ":" + PORT,
		WriteTimeout: 5 * time.Second,
		ReadTimeout:  5 * time.Second,
	}
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
