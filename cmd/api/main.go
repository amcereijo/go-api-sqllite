package main

import (
	"log"
	"net"
	"net/http"

	"github.com/joho/godotenv"

	"github.com/angel/go-api-sqlite/internal/database"
	grpcDelivery "github.com/angel/go-api-sqlite/internal/delivery/grpc"
	httpDelivery "github.com/angel/go-api-sqlite/internal/delivery/http"
	"github.com/angel/go-api-sqlite/internal/middleware"
	"github.com/angel/go-api-sqlite/internal/repositories/sqlite"
	"github.com/angel/go-api-sqlite/internal/usecases/feature"
	"github.com/angel/go-api-sqlite/internal/usecases/token"
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

	// Initialize repositories
	featureRepo := sqlite.NewFeatureRepository(db)
	tokenRepo := sqlite.NewTokenRepository(db)

	// Initialize use cases
	featureUseCase := feature.NewUseCase(featureRepo)
	tokenUseCase := token.NewUseCase(tokenRepo)

	// Initialize HTTP handlers
	healthHandler := httpDelivery.NewHealthHandler()
	featureHandler := httpDelivery.NewFeatureHandler(featureUseCase)
	tokenHandler := httpDelivery.NewTokenHandler(tokenUseCase)

	// Create router
	router := mux.NewRouter()

	// Apply CORS configuration
	router.Methods("OPTIONS").HandlerFunc(middleware.OptionsCors)
	router.Use(middleware.CorsMiddleware)

	// Public routes (no auth required)
	publicRouter := router.PathPrefix("/api").Subrouter()
	publicRouter.HandleFunc("/health", healthHandler.HealthCheck).Methods("GET")

	// Protected routes
	protectedRouter := router.PathPrefix("/api").Subrouter()
	protectedRouter.Use(middleware.AuthMiddleware)

	// Register routes
	featureHandler.RegisterRoutes(protectedRouter)
	tokenHandler.RegisterRoutes(protectedRouter)

	// Start gRPC server
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("Failed to listen for gRPC: %v", err)
	}

	// Create gRPC server with auth interceptor
	s := grpc.NewServer(
		grpc.UnaryInterceptor(middleware.GRPCAuthInterceptor()),
	)

	// Initialize and register gRPC service
	grpcServer := grpcDelivery.NewFeatureServer(featureUseCase)
	pb.RegisterFeatureServiceServer(s, grpcServer)

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
