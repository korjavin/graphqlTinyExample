# GraphQL Tiny Example

A simple GraphQL application demonstrating GraphQL with Go and PostgreSQL.

## Overview

This project is a simple example of how to implement a GraphQL API in Go, using a PostgreSQL database. It consists of:

1. A PostgreSQL database with tables for sellers, listings, purchases, and deliveries
2. A GraphQL server that serves data from the database, with filtering capabilities
3. A CLI client that can query the GraphQL server
4. Docker configurations for all components
5. Support for GraphQL mutations and real-time subscriptions

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
│   ├── events/            # Event system for subscriptions
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
- **Full CRUD operations** via queries and mutations
- **Real-time updates** with GraphQL subscriptions
- **Dockerized components** for easy deployment
- **CLI client** for interacting with the GraphQL API
- **GitHub Actions** for CI/CD pipeline

## GraphQL Schema

```graphql
# Main types
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

type Mutation {
  createListing(input: CreateListingInput!): Listing!
  createPurchase(input: CreatePurchaseInput!): Purchase!
  createDelivery(input: CreateDeliveryInput!): Delivery!
}

type Subscription {
  deliveryUpdated(purchaseId: ID): Delivery!
}

# Entity types with their relationships
type Seller { ... }
type Listing { ... }
type Purchase { ... }
type Delivery { ... }

# Filter and input types
input ListingFilter { ... }
input PurchaseFilter { ... }
input DeliveryFilter { ... }
input CreateListingInput { ... }
input CreatePurchaseInput { ... }
input CreateDeliveryInput { ... }
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

### Example Mutations

#### Create a New Listing
```graphql
mutation {
  createListing(input: {
    sellerId: "1",
    title: "New Gaming Laptop",
    description: "High performance gaming laptop with RTX 3080",
    price: 1299.99
  }) {
    id
    title
    price
  }
}
```

#### Create a Purchase
```graphql
mutation {
  createPurchase(input: {
    listingId: "5",
    price: 1299.99,
    bankTxId: "TX123456789",
    deliveryAddress: "123 Main St, Anytown, US 12345"
  }) {
    id
    price
    deliveryAddress
    createdAt
  }
}
```

#### Create a Delivery
```graphql
mutation {
  createDelivery(input: {
    purchaseId: "3",
    status: "PACKED"
  }) {
    id
    status
    timestamp
  }
}
```

### Example Subscriptions

#### Subscribe to Delivery Updates
```graphql
subscription {
  deliveryUpdated(purchaseId: "3") {
    id
    status
    timestamp
    purchase {
      id
      deliveryAddress
    }
  }
}
```

This subscription will provide real-time updates whenever a delivery status changes for the specified purchase ID. If no purchase ID is provided, it will subscribe to all delivery updates across the system.

## Real-time Capabilities

The application now supports real-time updates through GraphQL subscriptions:

1. **WebSocket Connection**: Clients can establish a WebSocket connection to subscribe to events
2. **Filtered Subscriptions**: Subscribe to specific purchase delivery updates or all updates
3. **Event-driven Architecture**: The system uses an event bus to manage and distribute events
4. **Low-latency Updates**: Receive instant notifications when delivery status changes

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
6. **Real-time capabilities**: Subscriptions enable real-time data updates
7. **Complete CRUD operations**: Full support for create, read, update, and delete operations through queries and mutations