package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/mascarock/gopayments/handler"
	"github.com/mascarock/gopayments/store"
)

func main() {
	dataPath := os.Getenv("GOPAYMENTS_DATA")
	if dataPath == "" {
		dataPath = "gopayments.data.json"
	}

	r := handler.NewRouter(store.NewPersistent(dataPath),
		handler.SecurityHeaders,
		handler.RateLimit,
		handler.BodyLimit,
		chimw.Logger,
		chimw.Recoverer,
		cors.Handler(cors.Options{
			AllowedOrigins:   []string{"http://localhost:5173", "http://127.0.0.1:5173"},
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Payment-Token"},
			ExposedHeaders:   []string{"X-Payment-Amount", "X-Payment-Currency", "X-Payment-Network", "X-Payment-Endpoint"},
			AllowCredentials: true,
			MaxAge:           300,
		}),
	)

	fmt.Println("GoPayments API running on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
