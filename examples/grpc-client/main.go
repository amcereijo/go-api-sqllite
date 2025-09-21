package main

import (
	"context"
	"log"
	"time"

	pb "github.com/angel/go-api-sqlite/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"
)

func main() {
	// Connect to gRPC server
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create client
	client := pb.NewFeatureServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create sample value
	value, err := structpb.NewValue(map[string]interface{}{
		"enabled": true,
		"config": map[string]interface{}{
			"timeout": 30,
			"retries": 3,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create value: %v", err)
	}

	// Create a feature
	log.Println("Creating feature...")
	feature, err := client.CreateFeature(ctx, &pb.CreateFeatureRequest{
		Name:       "Test Feature",
		Value:      value,
		ResourceId: "test-resource",
		Active:     true,
	})
	if err != nil {
		log.Fatalf("Failed to create feature: %v", err)
	}
	log.Printf("Feature created: %v\n", feature)

	// Get the feature
	log.Printf("Getting feature %s...\n", feature.Id)
	getResp, err := client.GetFeature(ctx, &pb.GetFeatureRequest{Id: feature.Id})
	if err != nil {
		log.Fatalf("Failed to get feature: %v", err)
	}
	log.Printf("Got feature: %v\n", getResp)

	// Update the feature
	log.Printf("Updating feature %s...\n", feature.Id)
	updateValue, err := structpb.NewValue(map[string]interface{}{
		"enabled": false,
		"config": map[string]interface{}{
			"timeout": 60,
			"retries": 5,
		},
	})
	if err != nil {
		log.Fatalf("Failed to create update value: %v", err)
	}

	updateResp, err := client.UpdateFeature(ctx, &pb.UpdateFeatureRequest{
		Id:         feature.Id,
		Name:       "Updated Feature",
		Value:      updateValue,
		ResourceId: "test-resource",
		Active:     false,
	})
	if err != nil {
		log.Fatalf("Failed to update feature: %v", err)
	}
	log.Printf("Updated feature: %v\n", updateResp)

	// List all features
	log.Println("Listing all features...")
	listResp, err := client.ListFeatures(ctx, &pb.ListFeaturesRequest{})
	if err != nil {
		log.Fatalf("Failed to list features: %v", err)
	}
	log.Printf("Found %d features\n", len(listResp.Features))

	// Delete the feature
	log.Printf("Deleting feature %s...\n", feature.Id)
	_, err = client.DeleteFeature(ctx, &pb.DeleteFeatureRequest{Id: feature.Id})
	if err != nil {
		log.Fatalf("Failed to delete feature: %v", err)
	}
	log.Println("Feature deleted successfully")
}
