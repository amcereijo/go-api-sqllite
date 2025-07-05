package grpc

import (
	"context"
	"database/sql"
	"time"

	"github.com/angel/go-api-sqlite/internal/models"
	pb "github.com/angel/go-api-sqlite/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ItemServer struct {
	pb.UnimplementedItemServiceServer
	db *sql.DB
}

func NewItemServer(db *sql.DB) *ItemServer {
	return &ItemServer{db: db}
}

func (s *ItemServer) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.Item, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}

	item := models.Item{
		ID:        uuid.New().String(),
		Name:      req.Name,
		Value:     req.Value,
		CreatedAt: time.Now(),
	}

	_, err := s.db.ExecContext(ctx,
		"INSERT INTO items (id, name, value, created_at) VALUES (?, ?, ?, ?)",
		item.ID, item.Name, item.Value, item.CreatedAt)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Item{
		Id:        item.ID,
		Name:      item.Name,
		Value:     item.Value,
		CreatedAt: timestamppb.New(item.CreatedAt),
	}, nil
}

func (s *ItemServer) GetItem(ctx context.Context, req *pb.GetItemRequest) (*pb.Item, error) {
	var item models.Item
	err := s.db.QueryRowContext(ctx,
		"SELECT id, name, value, created_at FROM items WHERE id = ?",
		req.Id).Scan(&item.ID, &item.Name, &item.Value, &item.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, status.Error(codes.NotFound, "item not found")
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Item{
		Id:        item.ID,
		Name:      item.Name,
		Value:     item.Value,
		CreatedAt: timestamppb.New(item.CreatedAt),
	}, nil
}

func (s *ItemServer) ListItems(ctx context.Context, req *pb.ListItemsRequest) (*pb.ListItemsResponse, error) {
	rows, err := s.db.QueryContext(ctx, "SELECT id, name, value, created_at FROM items")
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	defer rows.Close()

	var items []*pb.Item
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name, &item.Value, &item.CreatedAt); err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
		items = append(items, &pb.Item{
			Id:        item.ID,
			Name:      item.Name,
			Value:     item.Value,
			CreatedAt: timestamppb.New(item.CreatedAt),
		})
	}

	return &pb.ListItemsResponse{Items: items}, nil
}

func (s *ItemServer) UpdateItem(ctx context.Context, req *pb.UpdateItemRequest) (*pb.Item, error) {
	result, err := s.db.ExecContext(ctx,
		"UPDATE items SET name = ?, value = ? WHERE id = ?",
		req.Name, req.Value, req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "item not found")
	}

	// Get the updated item
	var item models.Item
	err = s.db.QueryRowContext(ctx,
		"SELECT id, name, value, created_at FROM items WHERE id = ?",
		req.Id).Scan(&item.ID, &item.Name, &item.Value, &item.CreatedAt)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.Item{
		Id:        item.ID,
		Name:      item.Name,
		Value:     item.Value,
		CreatedAt: timestamppb.New(item.CreatedAt),
	}, nil
}

func (s *ItemServer) DeleteItem(ctx context.Context, req *pb.DeleteItemRequest) (*pb.DeleteItemResponse, error) {
	result, err := s.db.ExecContext(ctx, "DELETE FROM items WHERE id = ?", req.Id)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	if rowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "item not found")
	}

	return &pb.DeleteItemResponse{Success: true}, nil
}
