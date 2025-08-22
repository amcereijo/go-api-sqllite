package main

import (
	"log"
	"net"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/angel/go-api-sqlite/internal/database"
	grpcserver "github.com/angel/go-api-sqlite/internal/grpc"
	"github.com/angel/go-api-sqlite/internal/handlers"
	"github.com/angel/go-api-sqlite/internal/middleware"
	pb "github.com/angel/go-api-sqlite/proto"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Error initializing database:", err)
	}
	defer db.Close()

	// Create router
	router := mux.NewRouter()

	// Apply CORS configuration
	router.Methods("OPTIONS").HandlerFunc(middleware.OptionsCors)
	router.Use(middleware.CorsMiddleware)

	// Initialize handlers
	h := handlers.NewHandler(db)

	// Public routes (no auth required)
	publicRouter := router.PathPrefix("/api").Subrouter()
	publicRouter.HandleFunc("/health", h.HealthCheck).Methods("GET")

	// Protected routes
	protectedRouter := router.PathPrefix("/api").Subrouter()
	protectedRouter.Use(middleware.AuthMiddleware)

	// Define protected routes
	protectedRouter.HandleFunc("/features", h.GetFeatures).Methods("GET", "OPTIONS")
	protectedRouter.HandleFunc("/features", h.CreateFeature).Methods("POST")
	protectedRouter.HandleFunc("/features/{id}", h.GetFeature).Methods("GET")
	protectedRouter.HandleFunc("/features/{id}", h.UpdateFeature).Methods("PUT")
	protectedRouter.HandleFunc("/features/{id}", h.DeleteFeature).Methods("DELETE")

	// API Token routes
	protectedRouter.HandleFunc("/tokens", h.CreateAPIToken).Methods("POST")
	protectedRouter.HandleFunc("/tokens", h.ListAPITokens).Methods("GET")
	protectedRouter.HandleFunc("/tokens/{id}", h.DeleteAPIToken).Methods("DELETE")

	// To apply for other endpoints
	//protectedRouter.Use(middleware.APITokenMiddleware(h))

	// Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}
	// Create gRPC server with auth interceptor
	s := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GRPCAuthInterceptor()),
	)
	pb.RegisterFeatureServiceServer(s, grpcserver.NewFeatureServer(db))

	// Start gRPC server in a goroutine
	go func() {
		log.Println("gRPC server starting on :50051")
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// Start HTTP server
	log.Println("HTTP server starting on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
