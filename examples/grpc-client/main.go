package main

import (
	"context"
	"log"
	"time"

	pb "github.com/angel/go-api-sqlite/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Connect to gRPC server
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	// Create client
	client := pb.NewItemServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create an item
	log.Println("Creating item...")
	item, err := client.CreateItem(ctx, &pb.CreateItemRequest{
		Name:  "Test Item",
		Value: 29.99,
	})
	if err != nil {
		log.Fatalf("Failed to create item: %v", err)
	}
	log.Printf("Created item: ID=%s, Name=%s, Value=%.2f\n", item.Id, item.Name, item.Value)

	// Get the item
	log.Println("\nGetting item...")
	getItem, err := client.GetItem(ctx, &pb.GetItemRequest{
		Id: item.Id,
	})
	if err != nil {
		log.Fatalf("Failed to get item: %v", err)
	}
	log.Printf("Got item: ID=%s, Name=%s, Value=%.2f\n", getItem.Id, getItem.Name, getItem.Value)

	// Update the item
	log.Println("\nUpdating item...")
	updatedItem, err := client.UpdateItem(ctx, &pb.UpdateItemRequest{
		Id:    item.Id,
		Name:  "Updated Test Item",
		Value: 39.99,
	})
	if err != nil {
		log.Fatalf("Failed to update item: %v", err)
	}
	log.Printf("Updated item: ID=%s, Name=%s, Value=%.2f\n", updatedItem.Id, updatedItem.Name, updatedItem.Value)

	// List all items
	log.Println("\nListing all items...")
	listResponse, err := client.ListItems(ctx, &pb.ListItemsRequest{})
	if err != nil {
		log.Fatalf("Failed to list items: %v", err)
	}
	log.Printf("Found %d items:\n", len(listResponse.Items))
	for _, it := range listResponse.Items {
		log.Printf("- ID=%s, Name=%s, Value=%.2f\n", it.Id, it.Name, it.Value)
	}

	// Delete the item
	log.Println("\nDeleting item...")
	deleteResponse, err := client.DeleteItem(ctx, &pb.DeleteItemRequest{
		Id: item.Id,
	})
	if err != nil {
		log.Fatalf("Failed to delete item: %v", err)
	}
	log.Printf("Delete success: %v\n", deleteResponse.Success)

	// Verify deletion by trying to get the item
	log.Println("\nVerifying deletion...")
	_, err = client.GetItem(ctx, &pb.GetItemRequest{
		Id: item.Id,
	})
	if err != nil {
		log.Printf("Item successfully deleted, got expected error: %v\n", err)
	} else {
		log.Fatal("Item still exists after deletion!")
	}
}
