package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ClerkClaims struct {
	jwt.RegisteredClaims
	AZP         string   `json:"azp"`
	Permissions []string `json:"permissions"`
}

// AuthMiddleware verifies the JWT token from Clerk
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			http.Error(w, "Unauthorized - No token provided", http.StatusUnauthorized)
			return
		}

		// Parse and verify the token
		claims := &ClerkClaims{}
		parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Get the public key from environment variable (you should set this from Clerk's JWKS endpoint)
			publicKey := os.Getenv("CLERK_JWT_PUBLIC_KEY")
			key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
			if err != nil {
				return nil, fmt.Errorf("error parsing public key: %v", err)
			}
			return key, nil
		})

		if err != nil {
			http.Error(w, "Unauthorized - Invalid token", http.StatusUnauthorized)
			return
		}

		if !parsedToken.Valid {
			http.Error(w, "Unauthorized - Invalid token", http.StatusUnauthorized)
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), "userId", claims.Subject)
		ctx = context.WithValue(ctx, "permissions", claims.Permissions)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GRPCAuthInterceptor provides authentication for gRPC services
func GRPCAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "Missing metadata")
		}

		token := extractTokenFromMetadata(md)
		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "Missing token")
		}

		// Parse and verify the token
		claims := &ClerkClaims{}
		parsedToken, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}

			// Get the public key from environment variable
			publicKey := os.Getenv("CLERK_JWT_PUBLIC_KEY")
			key, err := jwt.ParseRSAPublicKeyFromPEM([]byte(publicKey))
			if err != nil {
				return nil, status.Errorf(codes.Unauthenticated, "error parsing public key: %v", err)
			}
			return key, nil
		})

		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "Invalid token")
		}

		if !parsedToken.Valid {
			return nil, status.Error(codes.Unauthenticated, "Invalid token")
		}

		// Add claims to context
		newCtx := context.WithValue(ctx, "userId", claims.Subject)
		newCtx = context.WithValue(newCtx, "permissions", claims.Permissions)

		return handler(newCtx, req)
	}
}

func extractToken(r *http.Request) string {
	bearerToken := r.Header.Get("Authorization")
	if len(strings.Split(bearerToken, " ")) == 2 {
		return strings.Split(bearerToken, " ")[1]
	}
	return ""
}

func extractTokenFromMetadata(md metadata.MD) string {
	if values := md.Get("authorization"); len(values) > 0 {
		if len(strings.Split(values[0], " ")) == 2 {
			return strings.Split(values[0], " ")[1]
		}
	}
	return ""
}

// GetUserFromContext retrieves the user ID from the context
func GetUserFromContext(ctx context.Context) (string, error) {
	userId, ok := ctx.Value("userId").(string)
	if !ok {
		return "", fmt.Errorf("user not found in context")
	}
	return userId, nil
}
