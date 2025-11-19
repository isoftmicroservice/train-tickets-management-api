# Train Ticket Service - gRPC API

A gRPC-based service for purchasing and managing train tickets from London to France.

## Features

- Purchase tickets (public API)
- View ticket receipts (authenticated)
- View all seat allocations (admin only)
- Remove users from train (authenticated)
- Modify seat assignments (authenticated)

## Prerequisites

- Go 1.21+
- Protocol Buffer Compiler (`protoc`)
- Go plugins: `protoc-gen-go`, `protoc-gen-go-grpc`

## Quick Start

### Installation

```bash
# Install dependencies
go mod download

# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate protobuf code
make proto
```

### Run Server

```bash
go run ./cmd/server
```

Server runs on `localhost:50051`

### Run Client

```bash
# Purchase ticket
go run ./cmd/client purchase John Doe john@example.com

# View receipt (requires JWT)
go run ./cmd/client receipt <jwt_token>

# View allocations (admin, requires JWT)
go run ./cmd/client allocations <admin_jwt_token>

# Remove user (requires JWT)
go run ./cmd/client remove <jwt_token> [email]

# Modify seat (requires JWT)
go run ./cmd/client modify <jwt_token> <section> <seat_number> [email]
```

## API Endpoints

### 1. PurchaseTicket (Public)
Purchase a ticket with automatic seat assignment.

**Request:** `first_name`, `last_name`, `email`  
**Response:** Receipt with seat assignment

### 2. ViewUserReceipt (Authenticated)
View your own ticket receipt. Requires JWT.

**Request:** Empty (user from JWT)  
**Response:** Receipt

### 3. ViewAllocations (Admin Only)
View all seat allocations, optionally filtered by section.

**Request:** Optional `section` filter  
**Response:** List of allocations

### 4. RemoveUserFromTrain (Authenticated)
Remove a user from the train. User can remove themselves, admin can remove any user.

**Request:** Optional `email` (for admin)  
**Response:** Success message

### 5. ModifyUserSeat (Authenticated)
Modify seat assignment. User can modify own seat, admin can modify any seat.

**Request:** `section`, `seat_number`, optional `email` (for admin)  
**Response:** Updated receipt

## JWT Authentication

JWTs must include:
- `email` - User's email
- `first_name` - User's first name
- `last_name` - User's last name
- `role` - "user" or "admin" (defaults to "user")

**Note:** JWT signatures are NOT validated (as per requirements). Only parsing is performed.

Example JWT creation:
```go
claims := jwt.MapClaims{
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "role": "user",
}
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
tokenString, _ := token.SignedString([]byte("test-secret"))
```

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...
```

## Project Structure

```
cloudBees/
├── api/              # Proto definitions & generated code
├── cmd/
│   ├── server/       # gRPC server
│   └── client/       # CLI client
├── internal/
│   ├── service/      # Service implementation
│   ├── store/        # In-memory storage
│   ├── auth/         # JWT parsing
│   ├── model/        # Domain models
│   └── config/       # Constants
└── docs/             # API documentation
```

## Configuration

- Route: London → France
- Price: $20 per ticket
- Capacity: 20 seats (2 sections × 10 seats)

## Build Commands

```bash
make proto      # Generate protobuf code
make build      # Build server and client
make test       # Run tests
make run-server # Run server
make run-client # Run client
```

## Documentation

- [API Documentation](docs/api.md) - Detailed API reference
- `api/ticket.proto` - Proto service definition

