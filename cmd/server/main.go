package main

import (
	"encoding/json"
	"go-jwt-mysql-lab/internal/auth"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	"go-jwt-mysql-lab/internal/database"
	"go-jwt-mysql-lab/internal/user"

	httpapi "go-jwt-mysql-lab/internal/http"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	config := database.Config{
		Host:     os.Getenv("MYSQL_HOST"),
		Port:     os.Getenv("MYSQL_PORT"),
		User:     os.Getenv("MYSQL_USER"),
		Password: os.Getenv("MYSQL_PASSWORD"),
		Database: os.Getenv("MYSQL_DATABASE"),
	}

	db, err := database.NewMySQL(config)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userRepository := user.NewRepository(db)
	userService := user.NewService(userRepository)

	expirationMinutes, err := strconv.Atoi(
		os.Getenv("JWT_EXPIRATION_MINUTES"),
	)
	if err != nil {
		log.Fatal("invalid JWT_EXPIRATION_MINUTES")
	}

	jwtManager := auth.NewJWTManager(
		os.Getenv("JWT_SECRET"),
		os.Getenv("JWT_ISSUER"),
		time.Duration(expirationMinutes)*time.Minute,
	)

	handler := httpapi.NewHandler(userService, jwtManager)

	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/register", handler.Register)
	http.HandleFunc("/login", handler.Login)

	http.Handle(
		"/me",
		jwtManager.Middleware(http.HandlerFunc(handler.Me)),
	)

	http.Handle(
		"/protected",
		jwtManager.Middleware(http.HandlerFunc(handler.Protected)),
	)

	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server listening on :%s", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	response := map[string]string{
		"status": "ok",
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
