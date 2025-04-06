package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// GraphQL request structure
type graphQLRequest struct {
	Query         string                 `json:"query"`
	Variables     map[string]interface{} `json:"variables,omitempty"`
	OperationName string                 `json:"operationName,omitempty"`
}

// WebSocket message
type wsMessage struct {
	Type    string      `json:"type"`
	ID      string      `json:"id,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

// Command line flags
var (
	serverURL       string
	queryType       string
	id              int
	sellerId        int
	listingId       int
	minPrice        float64
	maxPrice        float64
	price           float64
	title           string
	description     string
	bankTxId        string
	deliveryAddress string
	statusFilter    string
	status          string
	fromDate        string
	toDate          string
	verbose         bool
)

func main() {
	// Setup logging
	log.SetPrefix("[GraphQL Client] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Parse command line flags
	serverURLEnv := os.Getenv("SERVER_URL")
	if serverURLEnv == "" {
		serverURLEnv = "http://localhost:8080/graphql"
	}

	flag.StringVar(&serverURL, "server", serverURLEnv, "GraphQL server URL")
	flag.StringVar(&queryType, "query", "", "Query/mutation type (sellers, seller, listings, listing, purchases, purchase, deliveries, delivery, create-listing, create-purchase, create-delivery, subscribe)")
	flag.IntVar(&id, "id", 0, "ID for specific item queries")
	flag.IntVar(&sellerId, "seller-id", 0, "Filter listings by seller ID or use as seller ID for creating listings")
	flag.IntVar(&listingId, "listing-id", 0, "Filter purchases by listing ID or use as listing ID for creating purchases")
	flag.Float64Var(&minPrice, "min-price", 0, "Filter listings by minimum price")
	flag.Float64Var(&maxPrice, "max-price", 0, "Filter listings by maximum price")
	flag.Float64Var(&price, "price", 0, "Price for creating listings or purchases")
	flag.StringVar(&title, "title", "", "Filter listings by title or use as title for creating listings")
	flag.StringVar(&description, "description", "", "Description for creating listings")
	flag.StringVar(&bankTxId, "bank-tx-id", "", "Bank transaction ID for creating purchases")
	flag.StringVar(&deliveryAddress, "delivery-address", "", "Delivery address for creating purchases")
	flag.StringVar(&statusFilter, "status", "", "Filter deliveries by status (PACKED, OUT_FOR_DELIVERY, DELIVERED, RESCHEDULED, CANCELED)")
	flag.StringVar(&status, "delivery-status", "", "Status for creating deliveries")
	flag.StringVar(&fromDate, "from", "", "Filter by start date (format: 2025-04-01T00:00:00Z)")
	flag.StringVar(&toDate, "to", "", "Filter by end date (format: 2025-04-01T00:00:00Z)")
	flag.BoolVar(&verbose, "v", false, "Verbose output")
	flag.Parse()

	log.Println("GraphQL client started")
	log.Printf("Server URL: %s", serverURL)

	// Check if query type is provided
	if queryType == "" {
		log.Println("No query type specified. Use -query flag with one of: sellers, seller, listings, listing, purchases, purchase, deliveries, delivery, create-listing, create-purchase, create-delivery, subscribe")
		flag.Usage()
		os.Exit(1)
	}

	// Build and execute the appropriate query
	var query string
	var variables map[string]interface{}

	switch queryType {
	// Existing query cases
	case "sellers":
		query = `
		query {
			sellers {
				id
				name
				address
			}
		}
		`
	case "seller":
		if id == 0 {
			log.Fatalf("Seller ID is required for seller query. Use -id flag.")
		}

		query = `
		query($id: ID!) {
			seller(id: $id) {
				id
				name
				address
				listings {
					id
					title
					price
				}
			}
		}
		`
		variables = map[string]interface{}{
			"id": strconv.Itoa(id),
		}
	case "listings":
		query = `
		query($filter: ListingFilter) {
			listings(filter: $filter) {
				id
				title
				description
				price
				seller {
					id
					name
				}
			}
		}
		`
		variables = buildListingFilter()
	case "listing":
		if id == 0 {
			log.Fatalf("Listing ID is required for listing query. Use -id flag.")
		}

		query = `
		query($id: ID!) {
			listing(id: $id) {
				id
				title
				description
				price
				seller {
					id
					name
					address
				}
				purchases {
					id
					price
					createdAt
				}
			}
		}
		`
		variables = map[string]interface{}{
			"id": strconv.Itoa(id),
		}
	case "purchases":
		query = `
		query($filter: PurchaseFilter) {
			purchases(filter: $filter) {
				id
				price
				bankTxId
				deliveryAddress
				createdAt
				listing {
					id
					title
					price
				}
			}
		}
		`
		variables = buildPurchaseFilter()
	case "purchase":
		if id == 0 {
			log.Fatalf("Purchase ID is required for purchase query. Use -id flag.")
		}

		query = `
		query($id: ID!) {
			purchase(id: $id) {
				id
				price
				bankTxId
				deliveryAddress
				createdAt
				listing {
					id
					title
					seller {
						id
						name
					}
				}
				deliveries {
					id
					timestamp
					status
				}
			}
		}
		`
		variables = map[string]interface{}{
			"id": strconv.Itoa(id),
		}
	case "deliveries":
		query = `
		query($filter: DeliveryFilter) {
			deliveries(filter: $filter) {
				id
				timestamp
				status
				purchase {
					id
					bankTxId
					listing {
						id
						title
					}
				}
			}
		}
		`
		variables = buildDeliveryFilter()

	case "delivery":
		if id == 0 {
			log.Fatalf("Delivery ID is required for delivery query. Use -id flag.")
		}

		query = `
		query($id: ID!) {
			delivery(id: $id) {
				id
				timestamp
				status
				purchase {
					id
					bankTxId
					deliveryAddress
					listing {
						id
						title
						seller {
							id
							name
						}
					}
				}
			}
		}
		`
		variables = map[string]interface{}{
			"id": strconv.Itoa(id),
		}

	// New mutation cases
	case "create-listing":
		if sellerId == 0 || title == "" || price == 0 {
			log.Fatalf("To create a listing, you must provide: -seller-id, -title, -price, and optionally -description")
		}

		query = `
		mutation($input: CreateListingInput!) {
			createListing(input: $input) {
				id
				title
				description
				price
				seller {
					id
					name
				}
			}
		}
		`
		variables = map[string]interface{}{
			"input": map[string]interface{}{
				"sellerId":    strconv.Itoa(sellerId),
				"title":       title,
				"description": description,
				"price":       price,
			},
		}

	case "create-purchase":
		if listingId == 0 || price == 0 || bankTxId == "" || deliveryAddress == "" {
			log.Fatalf("To create a purchase, you must provide: -listing-id, -price, -bank-tx-id, -delivery-address")
		}

		query = `
		mutation($input: CreatePurchaseInput!) {
			createPurchase(input: $input) {
				id
				price
				bankTxId
				deliveryAddress
				createdAt
				listing {
					id
					title
					seller {
						id
						name
					}
				}
			}
		}
		`
		variables = map[string]interface{}{
			"input": map[string]interface{}{
				"listingId":       strconv.Itoa(listingId),
				"price":           price,
				"bankTxId":        bankTxId,
				"deliveryAddress": deliveryAddress,
			},
		}

	case "create-delivery":
		if id == 0 || status == "" {
			log.Fatalf("To create a delivery, you must provide: -id (purchase ID), -delivery-status")
		}

		query = `
		mutation($input: CreateDeliveryInput!) {
			createDelivery(input: $input) {
				id
				timestamp
				status
				purchase {
					id
					bankTxId
					listing {
						id
						title
					}
				}
			}
		}
		`
		variables = map[string]interface{}{
			"input": map[string]interface{}{
				"purchaseId": strconv.Itoa(id),
				"status":     strings.ToUpper(status),
			},
		}

	case "subscribe":
		if id == 0 {
			log.Fatalf("Purchase ID is required for delivery subscription. Use -id flag.")
		}

		query = `
		subscription($purchaseId: ID!) {
			deliveryUpdated(purchaseId: $purchaseId) {
				id
				timestamp
				status
				purchase {
					id
					bankTxId
				}
			}
		}
		`
		variables = map[string]interface{}{
			"purchaseId": strconv.Itoa(id),
		}

		err := executeSubscription(query, variables)
		if err != nil {
			log.Fatalf("Failed to execute subscription: %v", err)
		}
		return

	default:
		log.Fatalf("Unknown query type: %s", queryType)
	}

	// Execute the GraphQL query
	startTime := time.Now()
	log.Printf("Executing %s...", queryType)
	if verbose {
		log.Printf("Query: %s", query)
		log.Printf("Variables: %+v", variables)
	}

	result, err := executeQuery(query, variables)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}

	elapsed := time.Since(startTime)

	// Pretty print the result
	fmt.Println("Query Result:")
	fmt.Println("=============")
	prettyJSON, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("Failed to format JSON: %v", err)
	}
	fmt.Println(string(prettyJSON))
	fmt.Println("=============")
	fmt.Printf("Executed in: %s\n", elapsed)
}

// buildListingFilter builds a filter for listings query based on command line flags
func buildListingFilter() map[string]interface{} {
	filter := make(map[string]interface{})
	filterVars := make(map[string]interface{})

	if sellerId > 0 {
		filterVars["sellerId"] = strconv.Itoa(sellerId)
	}

	if minPrice > 0 {
		filterVars["minPrice"] = minPrice
	}

	if maxPrice > 0 {
		filterVars["maxPrice"] = maxPrice
	}

	if title != "" {
		filterVars["title"] = title
	}

	if len(filterVars) > 0 {
		filter["filter"] = filterVars
	}

	return filter
}

// buildPurchaseFilter builds a filter for purchases query based on command line flags
func buildPurchaseFilter() map[string]interface{} {
	filter := make(map[string]interface{})
	filterVars := make(map[string]interface{})

	if listingId > 0 {
		filterVars["listingId"] = strconv.Itoa(listingId)
	}

	if bankTxId != "" {
		filterVars["bankTxId"] = bankTxId
	}

	if fromDate != "" {
		filterVars["fromDate"] = fromDate
	}

	if toDate != "" {
		filterVars["toDate"] = toDate
	}

	if len(filterVars) > 0 {
		filter["filter"] = filterVars
	}

	return filter
}

// buildDeliveryFilter builds a filter for deliveries query based on command line flags
func buildDeliveryFilter() map[string]interface{} {
	filter := make(map[string]interface{})
	filterVars := make(map[string]interface{})

	if id > 0 {
		filterVars["purchaseId"] = strconv.Itoa(id)
	}

	if statusFilter != "" {
		filterVars["status"] = strings.ToUpper(statusFilter)
	}

	if fromDate != "" {
		filterVars["fromDate"] = fromDate
	}

	if toDate != "" {
		filterVars["toDate"] = toDate
	}

	if len(filterVars) > 0 {
		filter["filter"] = filterVars
	}

	return filter
}

// executeQuery sends a GraphQL query to the server and returns the response
func executeQuery(query string, variables map[string]interface{}) (map[string]interface{}, error) {
	// Prepare the request
	reqBody, err := json.Marshal(graphQLRequest{
		Query:     query,
		Variables: variables,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create a new HTTP request
	req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Log the request details in verbose mode
	if verbose {
		log.Printf("Request URL: %s", serverURL)
		log.Printf("Request Body: %s", string(reqBody))
	}

	// Send the request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Log the response status code
	log.Printf("Response Status: %s", resp.Status)

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log the response body in verbose mode
	if verbose {
		log.Printf("Response Body: %s", string(body))
	}

	// Parse the response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Check for GraphQL errors
	if errors, ok := result["errors"]; ok {
		return nil, fmt.Errorf("GraphQL error: %v", errors)
	}

	return result, nil
}

// executeSubscription handles GraphQL subscriptions over WebSocket
func executeSubscription(query string, variables map[string]interface{}) error {
	// Convert HTTP URL to WebSocket URL
	wsURL := strings.Replace(serverURL, "http://", "ws://", 1)
	wsURL = strings.Replace(wsURL, "https://", "wss://", 1)
	wsURL = strings.Replace(wsURL, "/graphql", "/graphql/ws", 1)

	log.Printf("Connecting to WebSocket endpoint: %s", wsURL)

	// Connect to WebSocket
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	defer conn.Close()

	subscriptionID := uuid.New().String()

	// Send connection init message
	initMessage := wsMessage{Type: "connection_init"}
	if err := conn.WriteJSON(initMessage); err != nil {
		return fmt.Errorf("failed to send connection init: %w", err)
	}

	// Wait for connection ack
	var ackMessage wsMessage
	if err := conn.ReadJSON(&ackMessage); err != nil {
		return fmt.Errorf("failed to receive connection ack: %w", err)
	}

	if ackMessage.Type != "connection_ack" {
		return fmt.Errorf("expected connection_ack, got %s", ackMessage.Type)
	}

	log.Printf("Connection established, sending subscription request")

	// Send start subscription message
	startMessage := wsMessage{
		Type: "start",
		ID:   subscriptionID,
		Payload: graphQLRequest{
			Query:     query,
			Variables: variables,
		},
	}
	if err := conn.WriteJSON(startMessage); err != nil {
		return fmt.Errorf("failed to start subscription: %w", err)
	}

	log.Printf("Subscription started with ID: %s", subscriptionID)
	log.Printf("Listening for delivery updates (Press Ctrl+C to stop)...")

	// Handle incoming messages
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			var message wsMessage
			if err := conn.ReadJSON(&message); err != nil {
				log.Printf("Error reading WebSocket message: %v", err)
				return
			}

			switch message.Type {
			case "data":
				// Parse and display the subscription data
				if payload, ok := message.Payload.(map[string]interface{}); ok {
					if data, ok := payload["data"].(map[string]interface{}); ok {
						if update, ok := data["deliveryUpdated"].(map[string]interface{}); ok {
							fmt.Println("\n📦 Delivery Update Received:")
							fmt.Println("========================")
							prettyJSON, _ := json.MarshalIndent(update, "", "  ")
							fmt.Println(string(prettyJSON))
							fmt.Println("========================")
						}
					}
				} else {
					// Fallback for when type assertion fails
					prettyJSON, _ := json.MarshalIndent(message.Payload, "", "  ")
					fmt.Printf("\nReceived subscription data: %s\n", string(prettyJSON))
				}
			case "error":
				log.Printf("Subscription error: %v", message.Payload)
			case "complete":
				log.Printf("Subscription completed")
				return
			default:
				log.Printf("Received message of type: %s", message.Type)
			}
		}
	}()

	// Wait for interrupt signal to gracefully close the connection
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c

	log.Printf("Interrupted, closing subscription...")

	// Send stop subscription message
	stopMessage := wsMessage{Type: "stop", ID: subscriptionID}
	if err := conn.WriteJSON(stopMessage); err != nil {
		return fmt.Errorf("failed to stop subscription: %w", err)
	}

	// Wait for subscription to complete clean up
	select {
	case <-done:
	case <-time.After(1 * time.Second):
	}

	// Close connection
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(
		websocket.CloseNormalClosure, ""))

	return nil
}
