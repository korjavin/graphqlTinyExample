package graphql

// Schema contains the GraphQL schema definition
const Schema = `
schema {
  query: Query
}

type Query {
  # Seller queries
  seller(id: ID!): Seller
  sellers: [Seller!]!
  
  # Listing queries
  listing(id: ID!): Listing
  listings(filter: ListingFilter): [Listing!]!
  
  # Purchase queries
  purchase(id: ID!): Purchase
  purchases(filter: PurchaseFilter): [Purchase!]!
  
  # Delivery queries
  delivery(id: ID!): Delivery
  deliveries(filter: DeliveryFilter): [Delivery!]!
}

type Seller {
  id: ID!
  name: String!
  address: String!
  listings: [Listing!]!
}

type Listing {
  id: ID!
  seller: Seller!
  title: String!
  description: String!
  price: Float!
  purchases: [Purchase!]!
}

type Purchase {
  id: ID!
  listing: Listing!
  price: Float!
  bankTxId: String!
  deliveryAddress: String!
  createdAt: String!
  deliveries: [Delivery!]!
}

type Delivery {
  id: ID!
  purchase: Purchase!
  timestamp: String!
  status: DeliveryStatus!
}

enum DeliveryStatus {
  PACKED
  OUT_FOR_DELIVERY
  DELIVERED
  RESCHEDULED
  CANCELED
}

input ListingFilter {
  sellerId: ID
  minPrice: Float
  maxPrice: Float
  title: String
}

input PurchaseFilter {
  listingId: ID
  bankTxId: String
  fromDate: String
  toDate: String
}

input DeliveryFilter {
  purchaseId: ID
  status: DeliveryStatus
  fromDate: String
  toDate: String
}
`