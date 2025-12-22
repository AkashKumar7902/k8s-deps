package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	// Postgres driver
	_ "github.com/lib/pq"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	mongoClient *mongo.Client
	pgClient    *sql.DB

	// Config from Env
	MongoURI    = getEnv("MONGO_URI", "mongodb://myuser:mypassword@localhost:27017")
	PostgresURI = getEnv("POSTGRES_URI", "host=localhost port=5432 user=pguser password=pgpassword dbname=mydb sslmode=disable")
	Port        = ":8080"
)

func main() {
	log.Println("Application booting... waiting 20 seconds before starting server.")
	time.Sleep(20 * time.Second)

	// --- 1. Connect to Mongo ---
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	clientOptions := options.Client().ApplyURI(MongoURI)
	mongoClient, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Printf("Failed to connect to Mongo: %v", err)
	} else {
		if err := mongoClient.Ping(ctx, nil); err != nil {
			log.Printf("Mongo Ping failed: %v", err)
		} else {
			log.Println("Successfully connected to MongoDB")
		}
	}

	// --- 2. Connect to Postgres ---
	// We open the connection object, but Ping verifies connectivity
	pgClient, err = sql.Open("postgres", PostgresURI)
	if err != nil {
		log.Printf("Failed to open Postgres driver: %v", err)
	} else {
		if err := pgClient.Ping(); err != nil {
			log.Printf("Postgres Ping failed: %v", err)
		} else {
			log.Println("Successfully connected to Postgres")
		}
	}

	// --- 3. Define Endpoints ---
	http.HandleFunc("/hello", helloWorldHandler)
	http.HandleFunc("/noisy", noisyHandler)
	http.HandleFunc("/slow", slowHandler)
	http.HandleFunc("/mongo", mongoHandler)

	// New Endpoints
	http.HandleFunc("/external", externalApiHandler)
	http.HandleFunc("/postgres", postgresHandler)

	log.Printf("Server is ready and listening on %s", Port)
	if err := http.ListenAndServe(Port, nil); err != nil {
		log.Fatal(err)
	}
}

// --- Handlers ---

func helloWorldHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hello, World!"))
}

func noisyHandler(w http.ResponseWriter, r *http.Request) {
	type NoisyResponse struct {
		RequestID string `json:"request_id"`
		Nonce     int    `json:"nonce"`
		Timestamp int64  `json:"timestamp"`
	}
	resp := NoisyResponse{
		RequestID: fmt.Sprintf("req-%d", rand.Int63()),
		Nonce:     rand.Intn(999999),
		Timestamp: time.Now().UnixNano(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func slowHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received /slow request. Sleeping...")
	time.Sleep(35 * time.Second)
	w.Write([]byte("Finished processing after 35 seconds"))
}

func mongoHandler(w http.ResponseWriter, r *http.Request) {
	if mongoClient == nil {
		http.Error(w, "Mongo not initialized", 500)
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	dbs, err := mongoClient.ListDatabaseNames(ctx, bson.M{})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"source": "mongo", "dbs": dbs})
}

// New: Outgoing HTTPS call
func externalApiHandler(w http.ResponseWriter, r *http.Request) {
	// We will call a public JSON placeholder API
	url := "https://jsonplaceholder.typicode.com/todos/1"

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to call external API: %v", err), 502)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	w.Header().Set("Content-Type", "application/json")
	// Wrap the external response
	fmt.Fprintf(w, `{"source": "external_api", "url": "%s", "response": %s}`, url, string(body))
}

// New: Postgres Query
func postgresHandler(w http.ResponseWriter, r *http.Request) {
	if pgClient == nil {
		http.Error(w, "Postgres not initialized", 500)
		return
	}

	// Simple query to get the database version
	var version string
	err := pgClient.QueryRow("SELECT version()").Scan(&version)
	if err != nil {
		http.Error(w, fmt.Sprintf("Postgres query failed: %v", err), 500)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"source":  "postgres",
		"version": version,
	})
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
