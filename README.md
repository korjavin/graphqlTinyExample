# GraphQL Tiny Example

A simple GraphQL application demonstrating GraphQL with Go and PostgreSQL.

## Overview

This project is a simple example of how to implement a GraphQL API in Go, using a PostgreSQL database. It consists of:

1. A PostgreSQL database with tables for sellers, listings, purchases, and deliveries
2. A GraphQL server that serves data from the database, with filtering capabilities
3. A CLI client that can query the GraphQL server
4. Docker configurations for all components

## Project Structure

```
graphqlTinyExample/
├── .github/workflows/     # GitHub Actions workflows
├── cmd/
│   ├── client/            # CLI client implementation
│   │   └── Dockerfile
│   └── server/            # GraphQL server implementation
│       └── Dockerfile
├── migrations/            # Database migration files
│   ├── 01_schema.sql      # Schema definition
│   └── 02_fixtures.sql    # Test data
├── pkg/
│   ├── graphql/           # GraphQL schema and resolvers
│   ├── models/            # Data models
│   └── repository/        # Database operations
├── docker-compose.yaml    # Local development setup
├── docker-compose.gchr.yaml # Setup using pre-built images
├── Makefile               # Project management commands
└── README.md              # This file
```

## Features

- **GraphQL API** with filtering capabilities
- **Read-only data access** from the database
- **Dockerized components** for easy deployment
- **CLI client** for querying the GraphQL API
- **GitHub Actions** for CI/CD pipeline

## GraphQL Schema

```graphql
# Main query types
type Query {
  seller(id: ID!): Seller
  sellers: [Seller!]!
  listing(id: ID!): Listing
  listings(filter: ListingFilter): [Listing!]!
  purchase(id: ID!): Purchase
  purchases(filter: PurchaseFilter): [Purchase!]!
  delivery(id: ID!): Delivery
  deliveries(filter: DeliveryFilter): [Delivery!]!
}

# Entity types with their relationships
type Seller { ... }
type Listing { ... }
type Purchase { ... }
type Delivery { ... }

# Filter input types for querying data
input ListingFilter { ... }
input PurchaseFilter { ... }
input DeliveryFilter { ... }
```

## Setup & Usage

### Prerequisites

- Docker and Docker Compose
- Go 1.18+ (for local development)
- Make (optional, for using Makefile commands)

### Local Development

1. **Clone the repository:**
   ```bash
   git clone https://github.com/korjavin/graphqlTinyExample.git
   cd graphqlTinyExample
   ```

2. **Start the PostgreSQL database and set up the schema:**
   ```bash
   make db-setup
   ```

3. **Build and run the server:**
   ```bash
   make run-server
   ```

4. **In another terminal, run the client:**
   ```bash
   make run-client
   ```

### Using Docker Compose

1. **For local development with locally built images:**
   ```bash
   make up
   ```

2. **For using pre-built images from GHCR:**
   ```bash
   make up-ghcr
   ```

3. **Stop the containers:**
   ```bash
   make down
   ```

### CLI Client Usage Examples

```bash
# Get all sellers
./bin/client -query sellers

# Get a specific seller with ID 1
./bin/client -query seller -id 1

# Get listings filtered by seller ID and price range
./bin/client -query listings -seller-id 1 -min-price 50 -max-price 100

# Get deliveries with a specific status
./bin/client -query deliveries -status DELIVERED

# Get verbose output
./bin/client -query sellers -v
```

## GraphQL in Action

### Example Queries

#### Query All Sellers
```graphql
query {
  sellers {
    id
    name
    address
  }
}
```

#### Query Listings with Filter
```graphql
query {
  listings(filter: { minPrice: 50, maxPrice: 100, title: "Laptop" }) {
    id
    title
    price
    seller {
      name
    }
  }
}
```

#### Query Purchase with Related Data
```graphql
query {
  purchase(id: 1) {
    id
    price
    deliveryAddress
    listing {
      title
      seller {
        name
      }
    }
    deliveries {
      status
      timestamp
    }
  }
}
```

## Testing

Run the tests with:

```bash
make test
```

## Docker Images

Docker images are automatically built and published to GitHub Container Registry using GitHub Actions:

- Server image: `ghcr.io/korjavin/graphqltinyexample-server:latest`
- Client image: `ghcr.io/korjavin/graphqltinyexample-client:latest`

## Notes on GraphQL

GraphQL is a powerful query language for your API, providing several advantages:

1. **Client-specific data fetching**: Clients can request exactly what they need
2. **Single endpoint**: All data is accessible through a single endpoint
3. **Strong typing**: The schema provides a clear contract between client and server
4. **Introspection**: The API is self-documenting
5. **Efficient data loading**: Reduces over-fetching and under-fetching of data

While this example focuses on read operations, GraphQL also supports mutations (write operations) and subscriptions (real-time updates).