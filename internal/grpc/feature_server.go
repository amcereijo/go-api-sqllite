package grpc

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	pb "github.com/angel/go-api-sqlite/proto"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type FeatureServer struct {
	pb.UnimplementedFeatureServiceServer
	db *sql.DB
}

func NewFeatureServer(db *sql.DB) *FeatureServer {
	return &FeatureServer{db: db}
}

func (s *FeatureServer) CreateFeature(ctx context.Context, req *pb.CreateFeatureRequest) (*pb.Feature, error) {
	if req.Name == "" || req.ResourceId == "" {
		return nil, status.Error(codes.InvalidArgument, "name and resource_id are required")
	}

	id := uuid.New().String()
	now := time.Now()
	active := req.Active
	if !active {
		active = true // default to true if not specified
	}

	valueBytes, err := json.Marshal(req.Value.AsInterface())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal value: %v", err)
	}

	_, err = s.db.Exec(
		"INSERT INTO features (id, name, value, resource_id, active, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		id, req.Name, string(valueBytes), req.ResourceId, active, now,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create feature: %v", err)
	}

	createdAt := timestamppb.New(now)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert timestamp: %v", err)
	}

	return &pb.Feature{
		Id:         id,
		Name:       req.Name,
		Value:      req.Value,
		ResourceId: req.ResourceId,
		Active:     active,
		CreatedAt:  createdAt,
	}, nil
}

func (s *FeatureServer) GetFeature(ctx context.Context, req *pb.GetFeatureRequest) (*pb.Feature, error) {
	var feature pb.Feature
	var valueStr string
	var createdAt time.Time

	err := s.db.QueryRow(
		"SELECT id, name, value, resource_id, active, created_at FROM features WHERE id = ?",
		req.Id,
	).Scan(&feature.Id, &feature.Name, &valueStr, &feature.ResourceId, &feature.Active, &createdAt)

	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "feature not found")
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get feature: %v", err)
	}

	// Parse the value string back into a protobuf Value
	var value interface{}
	if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to unmarshal value: %v", err)
	}

	val, err := structpb.NewValue(value)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert value: %v", err)
	}
	feature.Value = val
	feature.CreatedAt = timestamppb.New(createdAt)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to convert timestamp: %v", err)
	}

	return &feature, nil
}

func (s *FeatureServer) ListFeatures(ctx context.Context, req *pb.ListFeaturesRequest) (*pb.ListFeaturesResponse, error) {
	var rows *sql.Rows
	var err error

	if req.ResourceId != "" {
		rows, err = s.db.Query(
			"SELECT id, name, value, resource_id, active, created_at FROM features WHERE resource_id = ?",
			req.ResourceId,
		)
	} else {
		rows, err = s.db.Query("SELECT id, name, value, resource_id, active, created_at FROM features")
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list features: %v", err)
	}
	defer rows.Close()

	var features []*pb.Feature
	for rows.Next() {
		var feature pb.Feature
		var valueStr string
		var createdAt time.Time

		err := rows.Scan(&feature.Id, &feature.Name, &valueStr, &feature.ResourceId, &feature.Active, &createdAt)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan feature: %v", err)
		}

		var value interface{}
		if err := json.Unmarshal([]byte(valueStr), &value); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to unmarshal value: %v", err)
		}

		val, err := structpb.NewValue(value)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to convert value: %v", err)
		}
		feature.Value = val
		feature.CreatedAt = timestamppb.New(createdAt)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to convert timestamp: %v", err)
		}

		features = append(features, &feature)
	}

	return &pb.ListFeaturesResponse{
		Features: features,
	}, nil
}

func (s *FeatureServer) UpdateFeature(ctx context.Context, req *pb.UpdateFeatureRequest) (*pb.Feature, error) {
	if req.Name == "" || req.ResourceId == "" {
		return nil, status.Error(codes.InvalidArgument, "name and resource_id are required")
	}

	valueBytes, err := json.Marshal(req.Value.AsInterface())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to marshal value: %v", err)
	}

	result, err := s.db.Exec(
		"UPDATE features SET name = ?, value = ?, resource_id = ?, active = ? WHERE id = ?",
		req.Name, string(valueBytes), req.ResourceId, req.Active, req.Id,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update feature: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "feature not found")
	}

	// Return the updated feature
	return s.GetFeature(ctx, &pb.GetFeatureRequest{Id: req.Id})
}

func (s *FeatureServer) DeleteFeature(ctx context.Context, req *pb.DeleteFeatureRequest) (*pb.DeleteFeatureResponse, error) {
	result, err := s.db.Exec("DELETE FROM features WHERE id = ?", req.Id)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete feature: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "feature not found")
	}

	return &pb.DeleteFeatureResponse{Success: true}, nil
}
