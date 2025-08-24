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
├── internal
│   ├── database
│   │   └── database.go
│   ├── delivery
│   │   ├── grpc
│   │   │   ├── feature_server.go
│   │   │   └── tests
│   │   │       ├── feature_test.go
│   │   │       └── mock_feature_usecase.go
│   │   └── http
│   │       ├── error_response.go
│   │       ├── feature_handler.go
│   │       ├── health_handler.go
│   │       ├── token_handler.go
│   │       └── tests
│   │           ├── create_feature_test.go
│   │           ├── feature_operations_test.go
│   │           ├── get_features_test.go
│   │           ├── health_test.go
│   │           ├── mock_feature_usecase.go
│   │           └── test_setup.go
│   ├── domain
│   │   ├── interfaces
│   │   │   └── repository.go
│   │   └── models
│   │       ├── api_token.go
│   │       └── feature.go
│   ├── grpc
│   ├── middleware
│   │   ├── auth.go
│   │   └── middleware.go
│   ├── models
│   │   ├── api_token.go
│   │   └── feature.go
│   ├── repositories
│   │   └── sqlite
│   │       ├── feature_repository.go
│   │       └── token_repository.go
│   └── usecases
│       ├── feature
│       │   └── feature_usecase.go
│       ├── interfaces
│       │   ├── feature.go
│       │   └── token.go
│       └── token
│           └── token_usecase.go
├── pkg
├── postman
│   └── go-sqlite-api.postman_collection.json
└── proto
    ├── feature_grpc.pb.go
    ├── feature.pb.go
    └── feature.proto
```

## Requirements

- Go 1.21 or higher
- SQLite3
- Protocol Buffers compiler (protoc)
- Clerk account for authentication (https://clerk.dev/)

## Getting Started

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Set up Clerk Authentication:
   - Create an account at https://clerk.dev/
   - Create a new application in Clerk
   - Get your JWT verification public key from Clerk Dashboard
   - Create a `.env` file in the project root with:
     ```bash
     CLERK_JWT_PUBLIC_KEY="-----BEGIN PUBLIC KEY-----
     YOUR_CLERK_PUBLIC_KEY_HERE
     -----END PUBLIC KEY-----"
     ```

4. Set up the database schema:
   The application will automatically create the required SQLite database with the necessary tables on first run.

5. Generate Protocol Buffer code (if modified):
   ```bash
   protoc -I=proto --go_out=. --go_opt=paths=source_relative \
          --go-grpc_out=. --go-grpc_opt=paths=source_relative \
          proto/feature.proto
   ```

5. Run the application:
   ```bash
   go run cmd/api/main.go
   ```

The HTTP server will start on port 8080 and the gRPC server on port 50051.

## Authentication

This API uses Clerk for authentication. All endpoints except `/api/health` require a valid JWT token from Clerk.

### Authentication Setup

1. Create a Clerk account and application
2. Get your JWT verification public key from Clerk Dashboard
3. Configure the `.env` file with your Clerk public key
4. Include the Clerk session token in all API requests:
   ```bash
   Authorization: Bearer <clerk_session_token>
   ```

### Public Endpoints
Only the health check endpoint is publicly accessible without authentication.

### Protected Endpoints
All other endpoints require either:
- A valid Clerk JWT token in the Authorization header
- A valid API token in the X-API-Token header

## API Endpoints

### Authentication

#### Create API Token
- `POST /api/tokens` - Create a new API token
  ```bash
  curl -X POST http://localhost:8080/api/tokens \
    -H "Authorization: Bearer YOUR_JWT_TOKEN" \
    -H "Content-Type: application/json" \
    -d '{
      "name": "My API Token"
    }'
  ```
  Response:
  ```json
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "My API Token",
    "createdAt": "2025-08-22T00:00:00Z",
    "created_by_uid": "user_123",
    "token": "YOUR_API_TOKEN"  // Only shown once at creation
  }
  ```

#### List API Tokens
- `GET /api/tokens` - List all API tokens for the authenticated user
  ```bash
  curl http://localhost:8080/api/tokens \
    -H "Authorization: Bearer YOUR_JWT_TOKEN"
  ```
  Response:
  ```json
  [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "My API Token",
      "createdAt": "2025-08-22T00:00:00Z",
      "lastUsedAt": "2025-08-22T01:00:00Z",
      "createdByUID": "user_123"
    }
  ]
  ```

#### Delete API Token
- `DELETE /api/tokens/{id}` - Delete an API token
  ```bash
  curl -X DELETE http://localhost:8080/api/tokens/123e4567-e89b-12d3-a456-426614174000 \
    -H "Authorization: Bearer YOUR_JWT_TOKEN"
  ```
  Response: `204 No Content`

### Using API Tokens
Once you have an API token, you can use it instead of JWT tokens to access protected endpoints:

```bash
curl http://localhost:8080/api/features \
  -H "X-API-Token: YOUR_API_TOKEN"
```

## API Endpoints

The API provides both REST (HTTP) and gRPC endpoints for all operations. All endpoints except `/api/health` require authentication.

### REST Endpoints

#### Health Check (Public)
- `GET /api/health` - Check if the API is running
  ```bash
  curl http://localhost:8080/api/health
  ```
  Response:
  ```json
  {"status": "healthy"}
  ```

### Features

#### Create Feature
- `POST /api/features` - Create a new feature flag
  ```bash
  curl -X POST http://localhost:8080/api/features \
    -H "Content-Type: application/json" \
    -d '{
      "name": "dark-mode",
      "value": true,
      "resourceId": "ui-settings",
      "active": true
    }'
  ```
  Response:
  ```json
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "dark-mode",
    "value": true,
    "resourceId": "ui-settings",
    "active": true,
    "createdAt": "2025-07-05T00:00:00Z"
  }
  ```

#### Get All Features
- `GET /api/features` - Retrieve all features
- `GET /api/features?resourceId=ui-settings` - Retrieve features by resource
  ```bash
  curl http://localhost:8080/api/features
  ```
  Response:
  ```json
  [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "name": "dark-mode",
      "value": true,
      "resourceId": "ui-settings",
      "active": true,
      "createdAt": "2025-07-05T00:00:00Z"
    }
  ]
  ```

#### Get Single Feature
- `GET /api/features/{id}` - Retrieve a specific feature
  ```bash
  curl http://localhost:8080/api/features/123e4567-e89b-12d3-a456-426614174000
  ```
  Response:
  ```json
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "dark-mode",
    "value": true,
    "resourceId": "ui-settings",
    "active": true,
    "createdAt": "2025-07-05T00:00:00Z"
  }
  ```

#### Update Feature
- `PUT /api/features/{id}` - Update an existing feature
  ```bash
  curl -X PUT http://localhost:8080/api/features/123e4567-e89b-12d3-a456-426614174000 \
    -H "Content-Type: application/json" \
    -d '{
      "name": "dark-mode",
      "value": false,
      "resourceId": "ui-settings",
      "active": false
    }'
  ```
  Response:
  ```json
  {
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "name": "dark-mode",
    "value": false,
    "resourceId": "ui-settings",
    "active": false,
    "createdAt": "2025-07-05T00:00:00Z"
  }
  ```

#### Delete Feature
- `DELETE /api/features/{id}` - Delete a feature
  ```bash
  curl -X DELETE http://localhost:8080/api/features/123e4567-e89b-12d3-a456-426614174000
  ```
  Response: `204 No Content`

### gRPC Service

The gRPC service is defined in `proto/feature.proto` and provides the following operations. All gRPC operations require authentication via the `authorization` metadata field with a valid Clerk JWT token:

```go
// Add authentication metadata to gRPC requests
md := metadata.New(map[string]string{
    "authorization": "Bearer " + clerkSessionToken,
})
ctx := metadata.NewOutgoingContext(context.Background(), md)
```

#### CreateFeature
```protobuf
rpc CreateFeature(CreateFeatureRequest) returns (Feature)
```
Example using the provided client:
```go
value, err := structpb.NewValue(map[string]interface{}{
    "enabled": true,
    "config": map[string]interface{}{
        "timeout": 30,
        "retries": 3,
    },
})
if err != nil {
    log.Fatal(err)
}

feature, err := client.CreateFeature(ctx, &pb.CreateFeatureRequest{
    Name:       "retry-config",
    Value:      value,
    ResourceId: "api-settings",
    Active:     true,
})
```

#### GetFeature
```protobuf
rpc GetFeature(GetFeatureRequest) returns (Feature)
```
Example:
```go
feature, err := client.GetFeature(ctx, &pb.GetFeatureRequest{
    Id: "123e4567-e89b-12d3-a456-426614174000",
})
```

#### ListFeatures
```protobuf
rpc ListFeatures(ListFeaturesRequest) returns (ListFeaturesResponse)
```
Example:
```go
// Get all features
features, err := client.ListFeatures(ctx, &pb.ListFeaturesRequest{})

// Get features for a specific resource
features, err := client.ListFeatures(ctx, &pb.ListFeaturesRequest{
    ResourceId: "api-settings",
})
```

#### UpdateFeature
```protobuf
rpc UpdateFeature(UpdateFeatureRequest) returns (Feature)
```
Example:
```go
value, err := structpb.NewValue(map[string]interface{}{
    "enabled": false,
    "config": map[string]interface{}{
        "timeout": 60,
        "retries": 5,
    },
})
if err != nil {
    log.Fatal(err)
}

feature, err := client.UpdateFeature(ctx, &pb.UpdateFeatureRequest{
    Id:         "123e4567-e89b-12d3-a456-426614174000",
    Name:       "retry-config",
    Value:      value,
    ResourceId: "api-settings",
    Active:     false,
})
```

#### DeleteFeature
```protobuf
rpc DeleteFeature(DeleteFeatureRequest) returns (DeleteFeatureResponse)
```
Example:
```go
response, err := client.DeleteFeature(ctx, &pb.DeleteFeatureRequest{
    Id: "123e4567-e89b-12d3-a456-426614174000",
})
```

### Response Behavior

The API follows consistent response patterns:

1. Empty Results:
   - When no records are found, endpoints return empty arrays (`[]`) instead of `null`
   - This applies to both REST and gRPC endpoints
   - Example:
     ```json
     // GET /api/features (when no features exist)
     []

     // GET /api/tokens (when no tokens exist)
     []
     ```

2. Error Responses:
   - HTTP 404 is returned only when requesting a specific resource by ID
   - List endpoints return empty arrays for no results

### Error Responses

Authentication Errors:
- `401 Unauthorized` - Missing or invalid authentication token
- `403 Forbidden` - Valid token but insufficient permissions

Other Errors:

- `400 Bad Request` - Invalid input (e.g., missing required fields)
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

### Postman Collection

A complete Postman collection for testing the API is available in the `postman` directory. To use it:

1. Open Postman
2. Click "Import" and select `postman/go-sqlite-api.postman_collection.json`
3. Create a new environment in Postman and add a variable:
   - `feature_id`: The ID of a feature you've created (you'll get this after creating your first feature)
   - `resource_id`: A resource identifier for testing resource-based filtering

Example Usage Flow:

1. **Health Check**
   - Send the "Health Check" request to verify the API is running

2. **Create Feature**
   - Send the "Create Feature" request with sample feature data
   - From the response, copy the `id` field

3. **Set Environment Variables**
   - In Postman, set the `feature_id` environment variable to the ID you copied
   - Set the `resource_id` variable for resource-based filtering tests

4. **Test Other Operations**
   - Now you can test Get, Update, and Delete operations using the saved ID
   - Try filtering features by resource using the saved resource_id

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

If you make changes to the protocol buffer definitions (`proto/feature.proto`), you'll need to regenerate the Go code:

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
          proto/feature.proto
   ```

### Development Workflow

When making changes to the codebase:

1. **Database Changes**
   - Update the database schema in `internal/database/database.go`
   - Update corresponding model in `internal/models/feature.go`
   - Run tests to verify changes: `go test ./internal/database/...`

2. **REST API Changes**
   - Update handlers in `internal/handlers/handlers.go`
   - Update corresponding tests in `internal/handlers/tests/`
   - Run tests: `go test ./internal/handlers/...`

3. **gRPC API Changes**
   - Update the protocol buffer definition in `proto/feature.proto`
   - Regenerate protocol buffer code (see above)
   - Update the gRPC server implementation in `internal/grpc/feature_server.go`
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
- `github.com/golang-jwt/jwt/v4` for JWT token validation
- `github.com/joho/godotenv` for environment variable management

## License

MIT
