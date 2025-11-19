# Take-Home Coding Exercise - Instructions

Please complete the following exercise using Golang and gRPC:

## Requirements

- Publish your code in a public GitHub repository and share a link we can access.
- Your code must compile. Unit tests are expected-full coverage isn't required, but it should not be zero.
- No persistence layer is needed; store all data in memory.
- Output may be printed via your gRPC server and gRPC client consoles.
- APIs should authenticate by parsing a JWT (no signature validation needed).

## Application Scenario

You are building a small gRPC-based service for purchasing and managing train tickets.

The user wants to board a train from London  France.
The ticket price is $20 (same for all seats and sections).

You must implement the following APIs (all via gRPC):

1. Public API - Purchase FFTicket

- Accept a ticket purchase request.
- Return a receipt containing:
  - From
  - To
  - User (first name, last name, email)
  - Price paid
- Automatically assign the user a seat.
  - Train has 2 sections: A and B
  - Each section has 10 seats
- Authenticated API - View User Receipt
  - Must read user info from a provided JWT in metadata.
- Authenticated Admin API - View Allocations
  - Admin can view all users and their assigned seats, filtered by section.
- Authenticated API - Remove User From Train
  - Accessible by the user or an admin.
- Authenticated API - Modify User's Seat
  - Accessible by the user or an admin.
