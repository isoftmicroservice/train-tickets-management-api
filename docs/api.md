# Train Ticket Service API Documentation

Generated from `api/ticket.proto`

## Service: TicketService

### PurchaseTicket

Public API to purchase a ticket. Automatically assigns a seat and returns a receipt.

**Request:** `PurchaseTicketRequest`
- `first_name` (string, required): User's first name
- `last_name` (string, required): User's last name  
- `email` (string, required): User's email address

**Response:** `PurchaseTicketResponse`
- `receipt` (Receipt): Ticket receipt with seat assignment

**Example:**
```bash
go run ./cmd/client purchase John Doe john@example.com
```

---

### ViewUserReceipt

Authenticated API to view user's own receipt. Reads user info from JWT in metadata.

**Request:** `ViewUserReceiptRequest`
- Empty request (user info comes from JWT)

**Response:** `ViewUserReceiptResponse`
- `receipt` (Receipt): User's ticket receipt

**Authentication:** Required (JWT in metadata)

**Example:**
```bash
go run ./cmd/client receipt <jwt_token>
```

---

### ViewAllocations

Admin API to view all seat allocations. Can be filtered by section.

**Request:** `ViewAllocationsRequest`
- `section` (string, optional): Filter by section ("A" or "B"). Empty means all sections

**Response:** `ViewAllocationsResponse`
- `allocations` (repeated Allocation): List of all seat allocations

**Authentication:** Required (Admin JWT)

**Example:**
```bash
go run ./cmd/client allocations <admin_jwt_token>
go run ./cmd/client allocations <admin_jwt_token> A
```

---

### RemoveUserFromTrain

Authenticated API to remove a user from the train. User can remove themselves, admin can remove any user.

**Request:** `RemoveUserFromTrainRequest`
- `email` (string, optional): Email of user to remove (for admin). If empty, removes the user from JWT

**Response:** `RemoveUserFromTrainResponse`
- `success` (bool): Operation success status
- `message` (string): Success/error message

**Authentication:** Required (JWT)

**Authorization:**
- User can remove themselves
- Admin can remove any user

**Example:**
```bash
go run ./cmd/client remove <jwt_token>
go run ./cmd/client remove <admin_jwt_token> user@example.com
```

---

### ModifyUserSeat

Authenticated API to modify a user's seat assignment. User can modify their own seat, admin can modify any user's seat.

**Request:** `ModifyUserSeatRequest`
- `email` (string, optional): Email of user to modify (for admin). If empty, modifies the user from JWT
- `section` (string, required): New section ("A" or "B")
- `seat_number` (int32, required): New seat number (1-10)

**Response:** `ModifyUserSeatResponse`
- `receipt` (Receipt): Updated ticket receipt

**Authentication:** Required (JWT)

**Authorization:**
- User can modify their own seat
- Admin can modify any user's seat

**Example:**
```bash
go run ./cmd/client modify <jwt_token> B 5
go run ./cmd/client modify <admin_jwt_token> A 3 user@example.com
```

---

## Message Types

### Receipt

Represents a ticket receipt.

- `from` (string): Departure city ("London")
- `to` (string): Destination city ("France")
- `user` (User): User information
- `price_paid` (int32): Price paid in cents ($20.00 = 2000)
- `seat` (Seat): Seat assignment

### User

Represents a user.

- `first_name` (string): User's first name
- `last_name` (string): User's last name
- `email` (string): User's email address

### Seat

Represents a seat assignment.

- `section` (string): Section identifier ("A" or "B")
- `seat_number` (int32): Seat number (1-10)

### Allocation

Represents a seat allocation.

- `section` (string): Section identifier
- `seat_number` (int32): Seat number
- `user` (User): User assigned to this seat

---

## Error Codes

- `InvalidArgument` (400): Invalid input parameters
- `Unauthenticated` (401): Missing or invalid JWT
- `PermissionDenied` (403): Insufficient permissions
- `NotFound` (404): Resource not found
- `AlreadyExists` (409): Resource already exists
- `ResourceExhausted` (429): Train is full

---

## Authentication

All authenticated endpoints require a JWT token in the `authorization` header:

```
authorization: Bearer <jwt_token>
```

### JWT Claims

```json
{
  "email": "user@example.com",
  "first_name": "John",
  "last_name": "Doe",
  "role": "user"  // or "admin"
}
```

**Note:** The service does NOT validate JWT signatures (as per requirements). It only parses the token to extract claims.

---

## gRPC Endpoints

- **Server:** `localhost:50051`
- **Service:** `ticket.TicketService`

### Full Method Names

- `/ticket.TicketService/PurchaseTicket`
- `/ticket.TicketService/ViewUserReceipt`
- `/ticket.TicketService/ViewAllocations`
- `/ticket.TicketService/RemoveUserFromTrain`
- `/ticket.TicketService/ModifyUserSeat`

