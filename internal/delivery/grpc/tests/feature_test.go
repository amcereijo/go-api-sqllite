package tests

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	grpcserver "github.com/angel/go-api-sqlite/internal/delivery/grpc"
	"github.com/angel/go-api-sqlite/internal/domain/models"
	pb "github.com/angel/go-api-sqlite/proto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/structpb"
)

const bufSize = 1024 * 1024

type testEnv struct {
	server      *grpc.Server
	client      pb.FeatureServiceClient
	mockUseCase *MockFeatureUseCase
	lis         *bufconn.Listener
}

func (env *testEnv) bufDialer(context.Context, string) (net.Conn, error) {
	return env.lis.Dial()
}

func newTestEnv(ctx context.Context) (*testEnv, error) {
	log.Println("Setting up test environment...")
	env := &testEnv{
		lis:         bufconn.Listen(bufSize),
		mockUseCase: new(MockFeatureUseCase),
		server:      grpc.NewServer(),
	}

	// No need to update the bufDialer since it's now a method on testEnv

	pb.RegisterFeatureServiceServer(env.server, grpcserver.NewFeatureServer(env.mockUseCase))

	ready := make(chan bool)
	go func() {
		log.Println("Starting gRPC server...")
		ready <- true
		if err := env.server.Serve(env.lis); err != nil && err != grpc.ErrServerStopped {
			log.Printf("Server exited with error: %v", err)
		}
	}()

	<-ready
	log.Println("gRPC server started, creating client...")

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(env.bufDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Printf("Failed to dial bufnet: %v", err)
		return nil, err
	}

	env.client = pb.NewFeatureServiceClient(conn)
	log.Println("Test environment setup completed.")

	return env, nil
}

func (env *testEnv) cleanup() {
	log.Println("Cleaning up test environment...")
	if env.server != nil {
		env.server.GracefulStop()
	}
	if env.lis != nil {
		env.lis.Close()
	}
	log.Println("Cleanup completed.")
}

func TestMain(m *testing.M) {
	log.SetOutput(os.Stdout) // Write logs to stdout for debugging
	os.Exit(m.Run())
}

func TestFeatureService(t *testing.T) {
	t.Log("Starting TestFeatureService")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	env, err := newTestEnv(ctx)
	require.NoError(t, err)
	defer env.cleanup()
	t.Log("Test setup completed")

	// Test Create Feature
	t.Run("Create Feature", func(t *testing.T) {
		tests := []struct {
			name    string
			req     *pb.CreateFeatureRequest
			mockFn  func(env *testEnv)
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
				mockFn: func(env *testEnv) {
					env.mockUseCase.On("CreateFeature", mock.Anything, mock.MatchedBy(func(f *models.Feature) bool {
						return f.Name == "Test Feature" && f.ResourceID == "resource-1" && f.Active == true
					})).Run(func(args mock.Arguments) {
						if f, ok := args.Get(1).(*models.Feature); ok {
							f.ID = "test-id-1"
						}
					}).Return(nil).Once()
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
				mockFn: func(env *testEnv) {
					env.mockUseCase.On("CreateFeature", mock.Anything, mock.MatchedBy(func(f *models.Feature) bool {
						return f.Name == "Number Feature" && f.ResourceID == "resource-1" && f.Active == true
					})).Run(func(args mock.Arguments) {
						if f, ok := args.Get(1).(*models.Feature); ok {
							f.ID = "test-id-2"
						}
					}).Return(nil).Once()
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
				mockFn: func(env *testEnv) {
					env.mockUseCase.On("CreateFeature", mock.Anything, mock.MatchedBy(func(f *models.Feature) bool {
						return f.Name == "" && f.ResourceID == "resource-1"
					})).Return(fmt.Errorf("validation error: name is required")).Once()
				},
				wantErr: codes.Internal,
			},
			{
				name: "Missing resourceId",
				req: &pb.CreateFeatureRequest{
					Name:  "Test Feature",
					Value: structpb.NewStringValue("test-value"),
				},
				mockFn: func(env *testEnv) {
					env.mockUseCase.On("CreateFeature", mock.Anything, mock.MatchedBy(func(f *models.Feature) bool {
						return f.Name == "Test Feature" && f.ResourceID == ""
					})).Return(fmt.Errorf("validation error: resourceId is required")).Once()
				},
				wantErr: codes.Internal,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.mockFn != nil {
					tt.mockFn(env)
				}

				resp, err := env.client.CreateFeature(ctx, tt.req)
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

				env.mockUseCase.AssertExpectations(t)
			})
		}
	})

	// Test Get Feature
	t.Run("Get Feature", func(t *testing.T) {
		testFeature := &models.Feature{
			ID:         "test-id-1",
			Name:       "Test Feature",
			ResourceID: "resource-1",
			Active:     true,
		}

		tests := []struct {
			name    string
			req     *pb.GetFeatureRequest
			mockFn  func(env *testEnv)
			want    *pb.Feature
			wantErr codes.Code
		}{
			{
				name: "Existing feature",
				req: &pb.GetFeatureRequest{
					Id: testFeature.ID,
				},
				mockFn: func(env *testEnv) {
					env.mockUseCase.On("GetFeatureByID", mock.Anything, testFeature.ID).Return(testFeature, nil).Once()
				},
				want: &pb.Feature{
					Id:         testFeature.ID,
					Name:       testFeature.Name,
					ResourceId: testFeature.ResourceID,
					Active:     testFeature.Active,
				},
				wantErr: codes.OK,
			},
			{
				name: "Non-existent feature",
				req: &pb.GetFeatureRequest{
					Id: "non-existent-id",
				},
				mockFn: func(env *testEnv) {
					env.mockUseCase.On("GetFeatureByID", mock.Anything, "non-existent-id").Return(nil, models.ErrFeatureNotFound).Once()
				},
				wantErr: codes.NotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.mockFn != nil {
					tt.mockFn(env)
				}

				resp, err := env.client.GetFeature(ctx, tt.req)
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
				assert.Equal(t, tt.want.ResourceId, resp.ResourceId)
				assert.Equal(t, tt.want.Active, resp.Active)

				env.mockUseCase.AssertExpectations(t)
			})
		}
	})

	// Test Update Feature
	t.Run("Update Feature", func(t *testing.T) {
		tests := []struct {
			name    string
			req     *pb.UpdateFeatureRequest
			mockFn  func(env *testEnv)
			want    *pb.Feature
			wantErr codes.Code
		}{
			{
				name: "Valid update",
				req: &pb.UpdateFeatureRequest{
					Id:         "test-id-1",
					Name:       "Updated Feature",
					Value:      structpb.NewStringValue("updated-value"),
					ResourceId: "resource-2",
					Active:     false,
				},
				mockFn: func(env *testEnv) {
					env.mockUseCase.On("UpdateFeature", mock.Anything, mock.MatchedBy(func(f *models.Feature) bool {
						return f.ID == "test-id-1" && f.Name == "Updated Feature" && f.ResourceID == "resource-2" && !f.Active
					})).Return(nil).Once()
				},
				want: &pb.Feature{
					Id:         "test-id-1",
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
				mockFn: func(env *testEnv) {
					env.mockUseCase.On("UpdateFeature", mock.Anything, mock.MatchedBy(func(f *models.Feature) bool {
						return f.ID == "non-existent-id"
					})).Return(models.ErrFeatureNotFound).Once()
				},
				wantErr: codes.NotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.mockFn != nil {
					tt.mockFn(env)
				}

				resp, err := env.client.UpdateFeature(ctx, tt.req)
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

				env.mockUseCase.AssertExpectations(t)
			})
		}
	})

	// Test Delete Feature
	t.Run("Delete Feature", func(t *testing.T) {
		tests := []struct {
			name    string
			req     *pb.DeleteFeatureRequest
			mockFn  func(env *testEnv)
			wantErr codes.Code
		}{
			{
				name: "Existing feature",
				req: &pb.DeleteFeatureRequest{
					Id: "test-id-1",
				},
				mockFn: func(env *testEnv) {
					env.mockUseCase.On("DeleteFeature", mock.Anything, "test-id-1").Return(nil).Once()
				},
				wantErr: codes.OK,
			},
			{
				name: "Non-existent feature",
				req: &pb.DeleteFeatureRequest{
					Id: "non-existent-id",
				},
				mockFn: func(env *testEnv) {
					env.mockUseCase.On("DeleteFeature", mock.Anything, "non-existent-id").Return(models.ErrFeatureNotFound).Once()
				},
				wantErr: codes.NotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.mockFn != nil {
					tt.mockFn(env)
				}

				resp, err := env.client.DeleteFeature(ctx, tt.req)
				if tt.wantErr != codes.OK {
					require.Error(t, err)
					st, ok := status.FromError(err)
					require.True(t, ok)
					assert.Equal(t, tt.wantErr, st.Code())
					return
				}

				require.NoError(t, err)
				assert.True(t, resp.Success)
				env.mockUseCase.AssertExpectations(t)
			})
		}
	})

	// Test List Features
	t.Run("List Features", func(t *testing.T) {
		mockFeatures := []*models.Feature{
			{
				ID:         "1",
				Name:       "Feature 1",
				ResourceID: "resource-1",
				Active:     true,
			},
			{
				ID:         "2",
				Name:       "Feature 2",
				ResourceID: "resource-1",
				Active:     false,
			},
			{
				ID:         "3",
				Name:       "Feature 3",
				ResourceID: "resource-2",
				Active:     true,
			},
		}

		env.mockUseCase.On("GetAllFeatures", mock.Anything).Return(mockFeatures, nil).Once()

		resp, err := env.client.ListFeatures(ctx, &pb.ListFeaturesRequest{})
		require.NoError(t, err)
		assert.Len(t, resp.Features, len(mockFeatures))

		for i, f := range resp.Features {
			assert.Equal(t, mockFeatures[i].ID, f.Id)
			assert.Equal(t, mockFeatures[i].Name, f.Name)
			assert.Equal(t, mockFeatures[i].ResourceID, f.ResourceId)
			assert.Equal(t, mockFeatures[i].Active, f.Active)
		}

		env.mockUseCase.AssertExpectations(t)
	})

	// Test Toggle Feature
	t.Run("Toggle Feature", func(t *testing.T) {
		tests := []struct {
			name    string
			req     *pb.ToggleFeatureRequest
			mockFn  func(env *testEnv)
			want    *pb.Feature
			wantErr codes.Code
		}{
			{
				name: "Successful toggle",
				req: &pb.ToggleFeatureRequest{
					Id:     "test-id-1",
					Active: false,
				},
				mockFn: func(env *testEnv) {
					env.mockUseCase.On("ToggleFeature", mock.Anything, "test-id-1", false).Return(nil).Once()
					env.mockUseCase.On("GetFeatureByID", mock.Anything, "test-id-1").Return(&models.Feature{
						ID:     "test-id-1",
						Name:   "Test Feature",
						Active: false,
					}, nil).Once()
				},
				want: &pb.Feature{
					Id:     "test-id-1",
					Name:   "Test Feature",
					Active: false,
				},
				wantErr: codes.OK,
			},
			{
				name: "Feature not found",
				req: &pb.ToggleFeatureRequest{
					Id:     "non-existent",
					Active: true,
				},
				mockFn: func(env *testEnv) {
					env.mockUseCase.On("ToggleFeature", mock.Anything, "non-existent", true).Return(models.ErrFeatureNotFound).Once()
				},
				wantErr: codes.NotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if tt.mockFn != nil {
					tt.mockFn(env)
				}

				resp, err := env.client.ToggleFeature(ctx, tt.req)
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
				assert.Equal(t, tt.want.Active, resp.Active)

				env.mockUseCase.AssertExpectations(t)
			})
		}
	})
}
