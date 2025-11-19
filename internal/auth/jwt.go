package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
)

var (
	ErrNoMetadata       = errors.New("no metadata provided")
	ErrNoAuthHeader     = errors.New("no authorization header")
	ErrInvalidToken     = errors.New("invalid token")
	ErrInvalidTokenFormat = errors.New("invalid token format")
)

type UserClaims struct {
	Email     string
	FirstName string
	LastName  string
	Role      string // "admin" or "user"
}

func (u *UserClaims) IsAdmin() bool {
	return u.Role == "admin"
}

func ExtractUserFromContext(ctx context.Context) (*UserClaims, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrNoMetadata
	}

	authHeaders := md.Get("authorization")
	if len(authHeaders) == 0 {
		return nil, ErrNoAuthHeader
	}

	authHeader := authHeaders[0]

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	if tokenString == "" {
		return nil, ErrInvalidTokenFormat
	}

	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	userClaims := &UserClaims{}

	if email, ok := claims["email"].(string); ok {
		userClaims.Email = email
	}

	if firstName, ok := claims["first_name"].(string); ok {
		userClaims.FirstName = firstName
	}

	if lastName, ok := claims["last_name"].(string); ok {
		userClaims.LastName = lastName
	}

	if role, ok := claims["role"].(string); ok {
		userClaims.Role = role
	} else {
		userClaims.Role = "user"
	}

	if userClaims.Email == "" {
		return nil, ErrInvalidToken
	}

	return userClaims, nil
}

