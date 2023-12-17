package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"

	// "math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	// "github.com/oklog/ulid/v2"

	hash "golang.org/x/crypto/bcrypt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	// "github.com/gorilla/csrf"
	// "github.com/gorilla/securecookie"
	// "github.com/gorilla/handlers"
)

type contextKey string

type User struct {
	ID           string `json:"id"`
	Nickname     string `json:"nickname"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
}

type Claims struct {
	UserRole   string `json:"uro"`
	TokenScope string `json:"scp"`
	jwt.RegisteredClaims
}

const (
	userKey    contextKey = "user"
	authIssuer string     = "Studopolis Authorization"
	dbURL      string     = "postgres://auth:auth@localhost:5432/postgres"
)

var connectionPool *pgxpool.Pool

// func generateULID() string {
// 	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
// 	return ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()
// }

func getECKeys() (*ecdsa.PrivateKey, *ecdsa.PublicKey) {
	// Public Key
	publicKeyData, err := os.ReadFile("public.pem")
	if err != nil {
		log.Println("Can not read the file")
	}

	publicBlock, _ := pem.Decode(publicKeyData)
	if publicBlock == nil || publicBlock.Type != "PUBLIC KEY" {
		log.Println("Failed to decode PEM block containing public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(publicBlock.Bytes)
	if err != nil {
		log.Println("Failed to parse EC public key")
	}

	// Pivate Key
	privateKeyData, err := os.ReadFile("private.pem")
	if err != nil {
		log.Println("Can not read the file")
	}

	privateBlock, _ := pem.Decode(privateKeyData)
	if privateBlock == nil || privateBlock.Type != "EC PRIVATE KEY" {
		log.Println("Failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParseECPrivateKey(privateBlock.Bytes)
	if err != nil {
		log.Println("Failed to parse EC private key")
	}

	return privateKey, publicKey.(*ecdsa.PublicKey)
}

// todo: func generateToken(user User, ttl time.Time) (string, error)
func generateToken(user *User) (string, error) {
	claims := &Claims{
		UserRole:   user.Role,
		TokenScope: "access", // todo: create custom type for token scope
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Minute)), // todo: add exp as an arg
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   fmt.Sprint(user.ID),
			Issuer:    authIssuer,
		},
	}

	privateKey, _ := getECKeys()
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(privateKey)
}

// func generatePasswordHash(password string) string {
// 	hash, err := hash.GenerateFromPassword([]byte(password), hash.MinCost)
// 	if err != nil {
// 		log.Println("Could not generate the password hash.", err)
// 	}

// 	return string(hash)
// }

// func checkHashedPassword(hashedPassword, password string) bool {
// 	return hash.CompareHashAndPassword([]byte(hashedPassword), []byte(password)) == nil
// }

func getUser(id string) (*User, error) {
	query := `
		select id, nickname, email, hash, role
		from public.users
		where id = $1;
	`
	user := &User{}
	row := connectionPool.QueryRow(context.Background(), query, id)
	err := row.Scan(&user.ID, &user.Nickname, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		log.Printf("Error quering the database: %v", err)
		return nil, err
	}

	return user, nil
}

func getUserByCredentials(email, password string) (*User, error) {
	query := `
		select id, nickname, email, hash, role
		from public.users
		where email = $1;
	`
	user := &User{}
	row := connectionPool.QueryRow(context.Background(), query, email)
	err := row.Scan(&user.ID, &user.Nickname, &user.Email, &user.PasswordHash, &user.Role)

	if err != nil {
		log.Printf("Error quering the database: %v", err)
		return nil, err
	}

	if err := hash.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.Println("Jopa")
		return nil, err
	}

	return user, nil
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var loginData struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := json.NewDecoder(r.Body).Decode(&loginData)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	var user = &User{}

	user, err = getUserByCredentials(loginData.Email, loginData.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	accessToken, err := generateToken(user)
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json") // is required?
	response := map[string]string{"token": accessToken}

	cookies := http.Cookie{
		Name:     "refresh",
		Value:    uuid.NewString(),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Now().Add(15 * time.Minute),
	}

	http.SetCookie(w, &cookies)
	json.NewEncoder(w).Encode(response)
}

func protectedHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(userKey).(*Claims).Subject
	user, err := getUser(userID)

	if err != nil {
		log.Printf("Error quering the user: %v", err)
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json") // is required?
	response := map[string]string{"message": fmt.Sprintf("Hello, %s! 🪲", user.Nickname)}

	json.NewEncoder(w).Encode(response)
}

func init() {
	// connection pool
	ctx := context.Background()

	poolConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		log.Fatalf("Unable to parse pool config: %v", err)
	}

	poolConfig.MaxConns = 10
	poolConfig.MinConns = 1

	connectionPool, err = pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("Unable to create pool: %v", err)
	}
}

func main() {
	// logger
	// logFileName := fmt.Sprintf("log/server_%d.log", time.Now().Unix())
	logFileName := "log/server.log"
	logFile, err := os.Create(logFileName)

	if err != nil {
		log.Fatal("Failed to create log file:", err)
	}
	defer logFile.Close()

	log.SetOutput(logFile)
	log.SetPrefix(fmt.Sprintf("[%s] ", uuid.NewString()))
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	// server
	rootRouter := mux.NewRouter().PathPrefix("/api/v1/").Subrouter()
	rootRouter.Use(loggingMiddleware)

	// todo: add CORS

	// corsMiddleware := handlers.CORS(
	// 	handlers.AllowedOrigins([]string{"https://example.com"}),
	// 	handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"}),
	// 	handlers.AllowedHeaders([]string{"Authorization", "Content-Type", "Origin"}),
	// 	handlers.AllowCredentials(),
	// 	handlers.MaxAge(3600),
	// )

	// rootRouter.Use(corsMiddleware)

	// public routers
	publicRouter := rootRouter.PathPrefix("/").Subrouter()
	publicRouter.HandleFunc("/login", loginHandler).Methods("POST")

	// protected routers
	protectedRouter := rootRouter.PathPrefix("/").Subrouter()
	protectedRouter.Use(jwtMiddleware)
	protectedRouter.HandleFunc("/protected", protectedHandler).Methods("GET")

	http.Handle("/", rootRouter)
	log.Println("Server is running on :8080")

	// todo: add certificate and key (DevOps)
	if err = http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		log.SetPrefix(fmt.Sprintf("[%s] ", uuid.NewString()))

		start := time.Now()
		log.Printf("Started %s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)
		log.Printf("Completed %s in %v", r.URL.Path, time.Since(start))
	}
	return http.HandlerFunc(fn)
}

func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenHeader := strings.TrimSpace(r.Header.Get("Authorization"))
		if tokenHeader == "" {
			http.Error(w, "Authorization header missing", http.StatusUnauthorized)
			return
		}

		var tokenString string
		if strings.HasPrefix(tokenHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(tokenHeader, "Bearer ")
		} else {
			tokenString = tokenHeader
		}

		requiredRole := r.Header.Get("X-Required-Role")
		if requiredRole == "" {
			http.Error(w, "User role missing", http.StatusForbidden)
			return
		}

		_, publicKey := getECKeys()
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return publicKey, nil
		})

		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if claims.UserRole != requiredRole {
			http.Error(w, "Access not granted", http.StatusForbidden)
			return
		}

		ctx := context.WithValue(r.Context(), userKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
