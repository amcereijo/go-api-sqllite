package tests

import (
	"context"
	"database/sql"
	"log"
	"net"
	"os"
	"testing"

	grpcserver "github.com/angel/go-api-sqlite/internal/grpc"
	pb "github.com/angel/go-api-sqlite/proto"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/structpb"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener
var client pb.FeatureServiceClient

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

	// Create the features table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS features (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			value TEXT NOT NULL,
			resource_id TEXT NOT NULL,
			active BOOLEAN DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// cleanup deletes all features from the database
func cleanup(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DELETE FROM features")
	require.NoError(t, err)
}

func TestMain(m *testing.M) {
	// Suppress log output during tests
	log.SetOutput(os.NewFile(0, os.DevNull))

	// Setup
	lis = bufconn.Listen(bufSize)
	db, err := setupTestServer()
	if err != nil {
		log.Fatalf("Failed to setup test server: %v", err)
	}
	defer db.Close()

	s := grpc.NewServer()
	pb.RegisterFeatureServiceServer(s, grpcserver.NewFeatureServer(db))
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	// Setup client
	conn, err := grpc.Dial("bufnet",
		grpc.WithContextDialer(bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()

	client = pb.NewFeatureServiceClient(conn)

	os.Exit(m.Run())
}

func TestFeatureService(t *testing.T) {
	ctx := context.Background()

	// Create a new database connection for each test to avoid sharing state
	db, err := setupTestServer()
	require.NoError(t, err)
	defer db.Close()

	// Test Create Feature
	t.Run("Create Feature", func(t *testing.T) {
		tests := []struct {
			name    string
			req     *pb.CreateFeatureRequest
			want    *pb.Feature
			wantErr codes.Code
		}{
			{
				name: "Valid feature with string value",
				req: &pb.CreateFeatureRequest{
					Name:       "Test Feature",
					Value:      structpb.NewStringValue("test-value"),
					ResourceId: "resource-1",
					Active:     true,
				},
				want: &pb.Feature{
					Name:       "Test Feature",
					Value:      structpb.NewStringValue("test-value"),
					ResourceId: "resource-1",
					Active:     true,
				},
				wantErr: codes.OK,
			},
			{
				name: "Valid feature with number value",
				req: &pb.CreateFeatureRequest{
					Name:       "Number Feature",
					Value:      structpb.NewNumberValue(42.0),
					ResourceId: "resource-1",
					Active:     true,
				},
				want: &pb.Feature{
					Name:       "Number Feature",
					Value:      structpb.NewNumberValue(42.0),
					ResourceId: "resource-1",
					Active:     true,
				},
				wantErr: codes.OK,
			},
			{
				name: "Missing name",
				req: &pb.CreateFeatureRequest{
					Value:      structpb.NewStringValue("test-value"),
					ResourceId: "resource-1",
				},
				wantErr: codes.InvalidArgument,
			},
			{
				name: "Missing resourceId",
				req: &pb.CreateFeatureRequest{
					Name:  "Test Feature",
					Value: structpb.NewStringValue("test-value"),
				},
				wantErr: codes.InvalidArgument,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := client.CreateFeature(ctx, tt.req)
				if tt.wantErr != codes.OK {
					require.Error(t, err)
					st, ok := status.FromError(err)
					require.True(t, ok)
					assert.Equal(t, tt.wantErr, st.Code())
					return
				}

				require.NoError(t, err)
				assert.NotEmpty(t, resp.Id)
				assert.Equal(t, tt.want.Name, resp.Name)
				assert.Equal(t, tt.want.Value.GetStringValue(), resp.Value.GetStringValue())
				assert.Equal(t, tt.want.ResourceId, resp.ResourceId)
				assert.Equal(t, tt.want.Active, resp.Active)
				assert.NotNil(t, resp.CreatedAt)
			})
		}
	})

	// Test Get Feature
	t.Run("Get Feature", func(t *testing.T) {
		// First create a feature
		createResp, err := client.CreateFeature(ctx, &pb.CreateFeatureRequest{
			Name:       "Test Feature",
			Value:      structpb.NewStringValue("test-value"),
			ResourceId: "resource-1",
			Active:     true,
		})
		require.NoError(t, err)

		tests := []struct {
			name    string
			req     *pb.GetFeatureRequest
			want    *pb.Feature
			wantErr codes.Code
		}{
			{
				name: "Existing feature",
				req: &pb.GetFeatureRequest{
					Id: createResp.Id,
				},
				want:    createResp,
				wantErr: codes.OK,
			},
			{
				name: "Non-existent feature",
				req: &pb.GetFeatureRequest{
					Id: "non-existent-id",
				},
				wantErr: codes.NotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := client.GetFeature(ctx, tt.req)
				if tt.wantErr != codes.OK {
					require.Error(t, err)
					st, ok := status.FromError(err)
					require.True(t, ok)
					assert.Equal(t, tt.wantErr, st.Code())
					return
				}

				require.NoError(t, err)
				assert.Equal(t, tt.want.Id, resp.Id)
				assert.Equal(t, tt.want.Name, resp.Name)
				assert.Equal(t, tt.want.Value.GetStringValue(), resp.Value.GetStringValue())
				assert.Equal(t, tt.want.ResourceId, resp.ResourceId)
				assert.Equal(t, tt.want.Active, resp.Active)
			})
		}
	})

	// Test Update Feature
	t.Run("Update Feature", func(t *testing.T) {
		// First create a feature
		createResp, err := client.CreateFeature(ctx, &pb.CreateFeatureRequest{
			Name:       "Original Feature",
			Value:      structpb.NewStringValue("original-value"),
			ResourceId: "resource-1",
			Active:     true,
		})
		require.NoError(t, err)

		tests := []struct {
			name    string
			req     *pb.UpdateFeatureRequest
			want    *pb.Feature
			wantErr codes.Code
		}{
			{
				name: "Valid update",
				req: &pb.UpdateFeatureRequest{
					Id:         createResp.Id,
					Name:       "Updated Feature",
					Value:      structpb.NewStringValue("updated-value"),
					ResourceId: "resource-2",
					Active:     false,
				},
				want: &pb.Feature{
					Id:         createResp.Id,
					Name:       "Updated Feature",
					Value:      structpb.NewStringValue("updated-value"),
					ResourceId: "resource-2",
					Active:     false,
				},
				wantErr: codes.OK,
			},
			{
				name: "Non-existent feature",
				req: &pb.UpdateFeatureRequest{
					Id:         "non-existent-id",
					Name:       "Updated Feature",
					Value:      structpb.NewStringValue("updated-value"),
					ResourceId: "resource-2",
					Active:     false,
				},
				wantErr: codes.NotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := client.UpdateFeature(ctx, tt.req)
				if tt.wantErr != codes.OK {
					require.Error(t, err)
					st, ok := status.FromError(err)
					require.True(t, ok)
					assert.Equal(t, tt.wantErr, st.Code())
					return
				}

				require.NoError(t, err)
				assert.Equal(t, tt.want.Id, resp.Id)
				assert.Equal(t, tt.want.Name, resp.Name)
				assert.Equal(t, tt.want.Value.GetStringValue(), resp.Value.GetStringValue())
				assert.Equal(t, tt.want.ResourceId, resp.ResourceId)
				assert.Equal(t, tt.want.Active, resp.Active)
			})
		}
	})

	// Test Delete Feature
	t.Run("Delete Feature", func(t *testing.T) {
		// First create a feature
		createResp, err := client.CreateFeature(ctx, &pb.CreateFeatureRequest{
			Name:       "Feature to Delete",
			Value:      structpb.NewStringValue("delete-me"),
			ResourceId: "resource-1",
			Active:     true,
		})
		require.NoError(t, err)

		tests := []struct {
			name    string
			req     *pb.DeleteFeatureRequest
			wantErr codes.Code
		}{
			{
				name: "Existing feature",
				req: &pb.DeleteFeatureRequest{
					Id: createResp.Id,
				},
				wantErr: codes.OK,
			},
			{
				name: "Non-existent feature",
				req: &pb.DeleteFeatureRequest{
					Id: "non-existent-id",
				},
				wantErr: codes.NotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := client.DeleteFeature(ctx, tt.req)
				if tt.wantErr != codes.OK {
					require.Error(t, err)
					st, ok := status.FromError(err)
					require.True(t, ok)
					assert.Equal(t, tt.wantErr, st.Code())
					return
				}

				require.NoError(t, err)
				assert.True(t, resp.Success)

				// Verify feature is deleted by trying to get it
				_, err = client.GetFeature(ctx, &pb.GetFeatureRequest{Id: tt.req.Id})
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, codes.NotFound, st.Code())
			})
		}
	})

	// Test List Features
	t.Run("List Features", func(t *testing.T) {
		// Create a new database connection for list features test
		db, err := setupTestServer()
		require.NoError(t, err)
		defer db.Close()

		// Set up a new gRPC server with the fresh database
		s := grpc.NewServer()
		pb.RegisterFeatureServiceServer(s, grpcserver.NewFeatureServer(db))
		lis := bufconn.Listen(bufSize)
		go func() {
			if err := s.Serve(lis); err != nil {
				log.Printf("Server exited with error: %v", err)
			}
		}()

		// Create a new client connection
		conn, err := grpc.Dial("bufnet",
			grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
				return lis.Dial()
			}),
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		require.NoError(t, err)
		defer conn.Close()

		client = pb.NewFeatureServiceClient(conn)

		// Create multiple features for testing
		features := []*pb.CreateFeatureRequest{
			{
				Name:       "Feature 1",
				Value:      structpb.NewStringValue("value-1"),
				ResourceId: "resource-1",
				Active:     true,
			},
			{
				Name:       "Feature 2",
				Value:      structpb.NewStringValue("value-2"),
				ResourceId: "resource-1",
				Active:     false,
			},
			{
				Name:       "Feature 3",
				Value:      structpb.NewStringValue("value-3"),
				ResourceId: "resource-2",
				Active:     true,
			},
		}

		for _, f := range features {
			_, err := client.CreateFeature(ctx, f)
			require.NoError(t, err)
		}

		tests := []struct {
			name      string
			req       *pb.ListFeaturesRequest
			wantCount int
			wantResID string
			wantErr   codes.Code
		}{
			{
				name:      "List all features",
				req:       &pb.ListFeaturesRequest{},
				wantCount: len(features),
				wantErr:   codes.OK,
			},
			{
				name: "List features by resource ID",
				req: &pb.ListFeaturesRequest{
					ResourceId: "resource-1",
				},
				wantCount: 2,
				wantResID: "resource-1",
				wantErr:   codes.OK,
			},
			{
				name: "List features with non-existent resource ID",
				req: &pb.ListFeaturesRequest{
					ResourceId: "non-existent",
				},
				wantCount: 0,
				wantResID: "non-existent",
				wantErr:   codes.OK,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				resp, err := client.ListFeatures(ctx, tt.req)
				if tt.wantErr != codes.OK {
					require.Error(t, err)
					st, ok := status.FromError(err)
					require.True(t, ok)
					assert.Equal(t, tt.wantErr, st.Code())
					return
				}

				require.NoError(t, err)
				assert.Len(t, resp.Features, tt.wantCount)
				if tt.wantResID != "" {
					for _, f := range resp.Features {
						assert.Equal(t, tt.wantResID, f.ResourceId)
					}
				}
			})
		}
	})
}
