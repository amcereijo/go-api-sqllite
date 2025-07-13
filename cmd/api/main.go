package main

import (
	"log"
	"net"
	"net/http"

	"github.com/angel/go-api-sqlite/internal/database"
	grpcserver "github.com/angel/go-api-sqlite/internal/grpc"
	"github.com/angel/go-api-sqlite/internal/handlers"
	pb "github.com/angel/go-api-sqlite/proto"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

func main() {
	// Initialize database
	db, err := database.InitDB()
	if err != nil {
		log.Fatal("Error initializing database:", err)
	}
	defer db.Close()

	// Create router
	router := mux.NewRouter()

	// Initialize handlers
	h := handlers.NewHandler(db)

	// Define routes
	router.HandleFunc("/api/health", h.HealthCheck).Methods("GET")
	router.HandleFunc("/api/features", h.GetFeatures).Methods("GET")
	router.HandleFunc("/api/features", h.CreateFeature).Methods("POST")
	router.HandleFunc("/api/features/{id}", h.GetFeature).Methods("GET")
	router.HandleFunc("/api/features/{id}", h.UpdateFeature).Methods("PUT")
	router.HandleFunc("/api/features/{id}", h.DeleteFeature).Methods("DELETE")

	// Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}
	s := grpc.NewServer()
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
