# Feature Flag Service with Go and SQLite

This is a dual REST and gRPC API built with Go and SQLite that provides feature flag management functionality. The service allows you to manage feature flags for different resources, supporting various value types (string, number, object) and flag states. The project follows a standard Go project layout and includes comprehensive CRUD operations with both HTTP and gRPC endpoints.

## Project Structure

```
.
├── cmd
│   └── api
│       └── main.go
├── examples
│   └── grpc-client
│       └── main.go
├── postman
│   └── go-sqlite-api.postman_collection.json
├── proto
│   ├── feature.proto
│   ├── feature.pb.go
│   └── feature_grpc.pb.go
└── internal
    ├── database
    │   └── database.go
    ├── grpc
    │   ├── feature_server.go
    │   └── tests
    │       └── feature_test.go
    ├── handlers
    │   ├── handlers.go
    │   └── tests
    │       ├── create_feature_test.go
    │       ├── feature_operations_test.go
    │       ├── get_features_test.go
    │       └── test_setup.go
    ├── middleware
    │   └── middleware.go
    └── models
        └── feature.go
```

## Requirements

- Go 1.21 or higher
- SQLite3
- Protocol Buffers compiler (protoc)

## Getting Started

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Generate Protocol Buffer code (if modified):
   ```bash
   protoc -I=proto --go_out=. --go_opt=paths=source_relative \
          --go-grpc_out=. --go-grpc_opt=paths=source_relative \
          proto/feature.proto
   ```

4. Run the application:
   ```bash
   go run cmd/api/main.go
   ```

The HTTP server will start on port 8080 and the gRPC server on port 50051.

## API Endpoints

The API provides both REST (HTTP) and gRPC endpoints for all operations.

### REST Endpoints

#### Health Check
- `GET /api/health` - Check if the API is running
  ```bash
  curl http://localhost:8080/api/health
  ```
  Response:
  ```json
  {"status": "healthy"}
  ```

### Items

#### Create Item
- `POST /api/items` - Create a new item
  ```bash
  curl -X POST http://localhost:8080/api/items \
    -H "Content-Type: application/json" \
    -d '{
      "name": "Test Item",
      "value": 29.99
    }'
  ```
  Response:
  ```json
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Test Item",
    "value": 29.99,
    "created_at": "2025-07-05T00:00:00Z"
  }
  ```

#### Get All Items
- `GET /api/items` - Retrieve all items
  ```bash
  curl http://localhost:8080/api/items
  ```
  Response:
  ```json
  [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "Test Item",
      "value": 29.99,
      "created_at": "2025-07-05T00:00:00Z"
    }
  ]
  ```

#### Get Single Item
- `GET /api/items/{id}` - Retrieve a specific item
  ```bash
  curl http://localhost:8080/api/items/123e4567-e89b-12d3-a456-426614174000
  ```
  Response:
  ```json
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Test Item",
    "value": 29.99,
    "created_at": "2025-07-05T00:00:00Z"
  }
  ```

#### Update Item
- `PUT /api/items/{id}` - Update an existing item
  ```bash
  curl -X PUT http://localhost:8080/api/items/123e4567-e89b-12d3-a456-426614174000 \
    -H "Content-Type: application/json" \
    -d '{
      "name": "Updated Item",
      "value": 39.99
    }'
  ```
  Response:
  ```json
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "Updated Item",
    "value": 39.99,
    "created_at": "2025-07-05T00:00:00Z"
  }
  ```

#### Delete Item
- `DELETE /api/items/{id}` - Delete an item
  ```bash
  curl -X DELETE http://localhost:8080/api/items/123e4567-e89b-12d3-a456-426614174000
  ```
  Response: `204 No Content`

### gRPC Service

The gRPC service is defined in `proto/item.proto` and provides the following operations:

#### CreateItem
```protobuf
rpc CreateItem(CreateItemRequest) returns (Item)
```
Example using the provided client:
```go
item, err := client.CreateItem(ctx, &pb.CreateItemRequest{
    Name:  "Test Item",
    Value: 29.99,
})
```

#### GetItem
```protobuf
rpc GetItem(GetItemRequest) returns (Item)
```
Example:
```go
item, err := client.GetItem(ctx, &pb.GetItemRequest{
    Id: "123e4567-e89b-12d3-a456-426614174000",
})
```

#### ListItems
```protobuf
rpc ListItems(ListItemsRequest) returns (ListItemsResponse)
```
Example:
```go
items, err := client.ListItems(ctx, &pb.ListItemsRequest{})
```

#### UpdateItem
```protobuf
rpc UpdateItem(UpdateItemRequest) returns (Item)
```
Example:
```go
item, err := client.UpdateItem(ctx, &pb.UpdateItemRequest{
    Id:    "123e4567-e89b-12d3-a456-426614174000",
    Name:  "Updated Item",
    Value: 39.99,
})
```

#### DeleteItem
```protobuf
rpc DeleteItem(DeleteItemRequest) returns (DeleteItemResponse)
```
Example:
```go
response, err := client.DeleteItem(ctx, &pb.DeleteItemRequest{
    Id: "123e4567-e89b-12d3-a456-426614174000",
})
```

### Example gRPC Client

A complete example gRPC client is provided in `examples/grpc-client/main.go`. To run it:

```bash
# First ensure the server is running
go run cmd/api/main.go

# In another terminal, run the example client
go run examples/grpc-client/main.go
```

### Error Responses

- `400 Bad Request` - Invalid input (e.g., missing required fields)
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

### Postman Collection

A complete Postman collection for testing the API is available in the `postman` directory. To use it:

1. Open Postman
2. Click "Import" and select `postman/go-sqlite-api.postman_collection.json`
3. Create a new environment in Postman and add a variable:
   - `item_id`: The ID of an item you've created (you'll get this after creating your first item)

Example Usage Flow:

1. **Health Check**
   - Send the "Health Check" request to verify the API is running

2. **Create Item**
   - Send the "Create Item" request
   - From the response, copy the `id` field

3. **Set Environment Variable**
   - In Postman, set the `item_id` environment variable to the ID you copied

4. **Test Other Operations**
   - Now you can test Get, Update, and Delete operations using the saved ID

The collection includes all API endpoints with proper headers, request bodies, and environment variables set up.

## Feature Flag Overview

The service provides a flexible feature flag system that supports:

- Dynamic feature values (strings, numbers, objects)
- Resource-based feature grouping
- Feature activation/deactivation
- REST and gRPC interfaces

### Feature Flag Model

Each feature flag consists of:

- `id`: Auto-generated unique identifier
- `name`: Feature name
- `value`: Feature value (supports string, number, or JSON object)
- `resourceId`: Resource identifier for grouping features
- `active`: Feature state (true/false, defaults to true)
- `createdAt`: Creation timestamp

### REST API Endpoints

#### Create Feature Flag
```http
POST /api/features
Content-Type: application/json

{
  "name": "dark-mode",
  "value": true,
  "resourceId": "ui-settings",
  "active": true
}
```

#### Get Features
```http
GET /api/features           # Get all features
GET /api/features?resourceId=ui-settings  # Get features for a specific resource
```

#### Get Feature by ID
```http
GET /api/features/{id}
```

#### Update Feature
```http
PUT /api/features/{id}
Content-Type: application/json

{
  "name": "dark-mode",
  "value": false,
  "resourceId": "ui-settings",
  "active": false
}
```

#### Delete Feature
```http
DELETE /api/features/{id}
```

### Example Value Types

The feature value field supports various types:

```json
// String value
{
  "name": "theme",
  "value": "dark",
  "resourceId": "ui-settings"
}

// Number value
{
  "name": "max-items",
  "value": 100,
  "resourceId": "pagination"
}

// Object value
{
  "name": "homepage-config",
  "value": {
    "showBanner": true,
    "layout": "grid",
    "columns": 3
  },
  "resourceId": "layout-settings"
}
```

### gRPC Service

The service also provides a gRPC interface with the following methods:

- `CreateFeature`
- `GetFeature`
- `ListFeatures`
- `UpdateFeature`
- `DeleteFeature`

See `proto/feature.proto` for the complete service definition.

## Building and Development

### Building the Project

1. Build the main application:
   ```bash
   go build -o api cmd/api/main.go
   ```

2. Build the example gRPC client:
   ```bash
   go build -o grpc-client examples/grpc-client/main.go
   ```

### Protocol Buffers

If you make changes to the protocol buffer definitions (`proto/item.proto`), you'll need to regenerate the Go code:

1. Install the Protocol Buffer compiler (protoc) if you haven't already:
   ```bash
   # macOS
   brew install protobuf

   # Ubuntu/Debian
   sudo apt-get install protobuf-compiler
   ```

2. Install Go Protocol Buffers plugins:
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

3. Regenerate the Protocol Buffer code:
   ```bash
   protoc --go_out=. --go_opt=paths=source_relative \
          --go-grpc_out=. --go-grpc_opt=paths=source_relative \
          proto/item.proto
   ```

### Development Workflow

When making changes to the codebase:

1. **Database Changes**
   - Update the database schema in `internal/database/database.go`
   - Update corresponding model in `internal/models/item.go`
   - Run tests to verify changes: `go test ./internal/database/...`

2. **REST API Changes**
   - Update handlers in `internal/handlers/handlers.go`
   - Update corresponding tests in `internal/handlers/tests/`
   - Run tests: `go test ./internal/handlers/...`

3. **gRPC API Changes**
   - Update the protocol buffer definition in `proto/item.proto`
   - Regenerate protocol buffer code (see above)
   - Update the gRPC server implementation in `internal/grpc/item_server.go`
   - Update corresponding tests in `internal/grpc/tests/`
   - Run tests: `go test ./internal/grpc/...`

4. **Client Changes**
   - Update the example gRPC client in `examples/grpc-client/main.go`
   - Test the client against a running server

## Testing

The project includes comprehensive test coverage for both REST and gRPC interfaces. Run the tests with:

```bash
go test ./...
```

Test coverage includes:
- Feature CRUD operations via REST API
- Feature CRUD operations via gRPC
- Input validation
- Error handling
- Resource-based feature filtering
- Value type handling (string, number, object)

The tests use an in-memory SQLite database to ensure isolation and fast execution.

## Development

This project uses:
- `github.com/gorilla/mux` for HTTP routing
- `github.com/mattn/go-sqlite3` for SQLite database operations
- `google.golang.org/grpc` for gRPC server and client
- `google.golang.org/protobuf` for Protocol Buffers support

## License

MIT
