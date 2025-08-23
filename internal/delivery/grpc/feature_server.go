package grpc

import (
	"context"
	"encoding/json"

	"github.com/angel/go-api-sqlite/internal/domain/models"
	"github.com/angel/go-api-sqlite/internal/usecases/interfaces"
	pb "github.com/angel/go-api-sqlite/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// FeatureServer implements the gRPC server
type FeatureServer struct {
	pb.UnimplementedFeatureServiceServer
	useCase interfaces.FeatureUseCase
}

// NewFeatureServer creates a new gRPC feature server
func NewFeatureServer(useCase interfaces.FeatureUseCase) *FeatureServer {
	return &FeatureServer{
		useCase: useCase,
	}
}

// toProtoFeature converts a domain feature to a proto feature
func toProtoFeature(f *models.Feature) (*pb.Feature, error) {
	if f == nil {
		return nil, nil
	}

	var val *structpb.Value
	var err error

	// Convert JSON RawMessage to structpb.Value
	if len(f.Value) > 0 {
		var v interface{}
		if err := json.Unmarshal(f.Value, &v); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to unmarshal value: %v", err)
		}
		val, err = structpb.NewValue(v)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to convert value: %v", err)
		}
	}

	return &pb.Feature{
		Id:         f.ID,
		Name:       f.Name,
		Value:      val,
		ResourceId: f.ResourceID,
		Active:     f.Active,
		CreatedAt:  timestamppb.New(f.CreatedAt),
	}, nil
}

// fromProtoValue converts a proto value to JSON RawMessage
func fromProtoValue(v *structpb.Value) (json.RawMessage, error) {
	if v == nil {
		return nil, nil
	}

	b, err := json.Marshal(v.AsInterface())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal value: %v", err)
	}
	return json.RawMessage(b), nil
}

// CreateFeature implements the CreateFeature RPC
func (s *FeatureServer) CreateFeature(ctx context.Context, req *pb.CreateFeatureRequest) (*pb.Feature, error) {
	value, err := fromProtoValue(req.Value)
	if err != nil {
		return nil, err
	}

	feature := &models.Feature{
		Name:       req.Name,
		Value:      value,
		ResourceID: req.ResourceId,
		Active:     req.Active,
	}

	if err := s.useCase.CreateFeature(ctx, feature); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create feature: %v", err)
	}

	return toProtoFeature(feature)
}

// GetFeature implements the GetFeature RPC
func (s *FeatureServer) GetFeature(ctx context.Context, req *pb.GetFeatureRequest) (*pb.Feature, error) {
	feature, err := s.useCase.GetFeatureByID(ctx, req.Id)
	if err != nil {
		if err == models.ErrFeatureNotFound {
			return nil, status.Error(codes.NotFound, "feature not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get feature: %v", err)
	}

	return toProtoFeature(feature)
}

// ListFeatures implements the ListFeatures RPC
func (s *FeatureServer) ListFeatures(ctx context.Context, req *pb.ListFeaturesRequest) (*pb.ListFeaturesResponse, error) {
	features, err := s.useCase.GetAllFeatures(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list features: %v", err)
	}

	protoFeatures := make([]*pb.Feature, len(features))
	for i, f := range features {
		pf, err := toProtoFeature(f)
		if err != nil {
			return nil, err
		}
		protoFeatures[i] = pf
	}

	return &pb.ListFeaturesResponse{
		Features: protoFeatures,
	}, nil
}

// UpdateFeature implements the UpdateFeature RPC
func (s *FeatureServer) UpdateFeature(ctx context.Context, req *pb.UpdateFeatureRequest) (*pb.Feature, error) {
	value, err := fromProtoValue(req.Value)
	if err != nil {
		return nil, err
	}

	feature := &models.Feature{
		ID:         req.Id,
		Name:       req.Name,
		Value:      value,
		ResourceID: req.ResourceId,
		Active:     req.Active,
	}

	if err := s.useCase.UpdateFeature(ctx, feature); err != nil {
		if err == models.ErrFeatureNotFound {
			return nil, status.Error(codes.NotFound, "feature not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to update feature: %v", err)
	}

	return toProtoFeature(feature)
}

// DeleteFeature implements the DeleteFeature RPC
func (s *FeatureServer) DeleteFeature(ctx context.Context, req *pb.DeleteFeatureRequest) (*pb.DeleteFeatureResponse, error) {
	if err := s.useCase.DeleteFeature(ctx, req.Id); err != nil {
		if err == models.ErrFeatureNotFound {
			return nil, status.Error(codes.NotFound, "feature not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to delete feature: %v", err)
	}

	return &pb.DeleteFeatureResponse{Success: true}, nil
}

// ToggleFeature implements the ToggleFeature RPC
func (s *FeatureServer) ToggleFeature(ctx context.Context, req *pb.ToggleFeatureRequest) (*pb.Feature, error) {
	if err := s.useCase.ToggleFeature(ctx, req.Id, req.Active); err != nil {
		if err == models.ErrFeatureNotFound {
			return nil, status.Error(codes.NotFound, "feature not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to toggle feature: %v", err)
	}

	feature, err := s.useCase.GetFeatureByID(ctx, req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get updated feature: %v", err)
	}

	return toProtoFeature(feature)
}
