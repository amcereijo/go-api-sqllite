package tests

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"testing"

	"github.com/angel/go-api-sqlite/internal/grpc"
	pb "github.com/angel/go-api-sqlite/proto"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	grpclib "google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener
var client pb.ItemServiceClient

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

// setupTestServer creates an in-memory SQLite database for testing
func setupTestServer() (*sql.DB, error) {
	// Use an in-memory SQLite database for testing
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return nil, err
	}

	// Create the items table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS items (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			value REAL NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func TestMain(m *testing.M) {
	// Suppress log output during tests
	log.SetOutput(os.NewFile(0, os.DevNull))

	// Set up the gRPC server with bufconn listener
	lis = bufconn.Listen(bufSize)
	s := grpclib.NewServer()
	db, err := setupTestServer()
	if err != nil {
		log.Fatalf("Failed to setup test database: %v", err)
	}

	pb.RegisterItemServiceServer(s, grpc.NewItemServer(db))
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve test server: %v", err)
		}
	}()

	// Set up the test client
	conn, err := grpclib.Dial("bufnet", grpclib.WithContextDialer(bufDialer), grpclib.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client = pb.NewItemServiceClient(conn)

	// Run the tests
	exitCode := m.Run()

	// Clean up
	db.Close()
	s.Stop()
	os.Exit(exitCode)
}

func TestCreateItem(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		request *pb.CreateItemRequest
		wantErr bool
	}{
		{
			name: "Valid item",
			request: &pb.CreateItemRequest{
				Name:  "Test Item",
				Value: 29.99,
			},
			wantErr: false,
		},
		{
			name: "Missing name",
			request: &pb.CreateItemRequest{
				Value: 29.99,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := client.CreateItem(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEmpty(t, response.Id)
			assert.Equal(t, tt.request.Name, response.Name)
			assert.Equal(t, tt.request.Value, response.Value)
			assert.NotNil(t, response.CreatedAt)
		})
	}
}

func TestGetItem(t *testing.T) {
	ctx := context.Background()

	// First create an item
	createResp, err := client.CreateItem(ctx, &pb.CreateItemRequest{
		Name:  "Test Item",
		Value: 29.99,
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		request *pb.GetItemRequest
		wantErr bool
	}{
		{
			name: "Existing item",
			request: &pb.GetItemRequest{
				Id: createResp.Id,
			},
			wantErr: false,
		},
		{
			name: "Non-existent item",
			request: &pb.GetItemRequest{
				Id: "non-existent-id",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := client.GetItem(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, createResp.Id, response.Id)
			assert.Equal(t, createResp.Name, response.Name)
			assert.Equal(t, createResp.Value, response.Value)
		})
	}
}

func TestListItems(t *testing.T) {
	ctx := context.Background()

	// Create a few items
	items := []struct {
		name  string
		value float64
	}{
		{"Item 1", 29.99},
		{"Item 2", 39.99},
		{"Item 3", 49.99},
	}

	for _, item := range items {
		_, err := client.CreateItem(ctx, &pb.CreateItemRequest{
			Name:  item.name,
			Value: item.value,
		})
		require.NoError(t, err)
	}

	response, err := client.ListItems(ctx, &pb.ListItemsRequest{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(response.Items), len(items))
}

func TestUpdateItem(t *testing.T) {
	ctx := context.Background()

	// First create an item
	createResp, err := client.CreateItem(ctx, &pb.CreateItemRequest{
		Name:  "Test Item",
		Value: 29.99,
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		request *pb.UpdateItemRequest
		wantErr bool
	}{
		{
			name: "Valid update",
			request: &pb.UpdateItemRequest{
				Id:    createResp.Id,
				Name:  "Updated Item",
				Value: 39.99,
			},
			wantErr: false,
		},
		{
			name: "Non-existent item",
			request: &pb.UpdateItemRequest{
				Id:    "non-existent-id",
				Name:  "Updated Item",
				Value: 39.99,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := client.UpdateItem(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.request.Id, response.Id)
			assert.Equal(t, tt.request.Name, response.Name)
			assert.Equal(t, tt.request.Value, response.Value)
		})
	}
}

func TestDeleteItem(t *testing.T) {
	ctx := context.Background()

	// First create an item
	createResp, err := client.CreateItem(ctx, &pb.CreateItemRequest{
		Name:  "Test Item",
		Value: 29.99,
	})
	require.NoError(t, err)

	tests := []struct {
		name    string
		request *pb.DeleteItemRequest
		wantErr bool
	}{
		{
			name: "Existing item",
			request: &pb.DeleteItemRequest{
				Id: createResp.Id,
			},
			wantErr: false,
		},
		{
			name: "Non-existent item",
			request: &pb.DeleteItemRequest{
				Id: "non-existent-id",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := client.DeleteItem(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.True(t, response.Success)

			// Verify deletion
			_, err = client.GetItem(ctx, &pb.GetItemRequest{Id: tt.request.Id})
			assert.Error(t, err)
		})
	}
}
