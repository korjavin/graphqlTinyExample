package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
	"github.com/graph-gophers/graphql-go/relay"
	_ "github.com/lib/pq"

	"github.com/korjavin/graphqlTinyExample/pkg/graphql"
	"github.com/korjavin/graphqlTinyExample/pkg/models"
	"github.com/korjavin/graphqlTinyExample/pkg/repository"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func main() {
	log.Println("Starting GraphQL server...")

	// Get database configuration from environment variables
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "graphql_example")

	// Connect to the database
	db, err := models.NewDB(dbHost, dbPort, dbUser, dbPassword, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Create repository and resolver
	repo := repository.NewRepository(db)
	resolver := graphql.NewResolver(repo)

	// Create GraphQL schema
	schema, err := graphql.GetSchema(resolver)
	if err != nil {
		log.Fatalf("Failed to create GraphQL schema: %v", err)
	}

	// Set up HTTP handler for regular GraphQL queries and mutations
	http.Handle("/graphql", corsMiddleware(&relay.Handler{Schema: schema}))

	// Set up WebSocket handler for GraphQL subscriptions
	http.HandleFunc("/graphql/ws", func(w http.ResponseWriter, r *http.Request) {
		// Upgrade HTTP connection to WebSocket
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("Failed to upgrade connection to WebSocket: %v", err)
			return
		}
		defer conn.Close()

		// Log the new WebSocket connection
		log.Printf("[WS] New WebSocket connection from %s", r.RemoteAddr)

		// Handle subscription protocol
		handleWebSocketConnection(conn, resolver)
	})

	// Serve GraphQL Playground for interactive API exploration
	http.HandleFunc("/", playgroundHandler)

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("Server started at http://localhost:%s/", port)
	log.Printf("GraphQL HTTP endpoint: http://localhost:%s/graphql", port)
	log.Printf("GraphQL WebSocket endpoint: http://localhost:%s/graphql/ws", port)
	log.Printf("GraphQL Playground: http://localhost:%s/", port)

	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// handleWebSocketConnection manages the WebSocket connection for GraphQL subscriptions
func handleWebSocketConnection(conn *websocket.Conn, resolver *graphql.Resolver) {
	// Map of active subscriptions, keyed by subscription ID
	subscriptions := make(map[string]chan struct{})
	defer func() {
		// Clean up all subscriptions when connection closes
		for id, quit := range subscriptions {
			close(quit)
			log.Printf("[WS] Closing subscription %s", id)
		}
	}()

	// Process WebSocket messages
	for {
		// Read message from WebSocket
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("[WS] Error reading message: %v", err)
			break
		}

		// Parse the message
		var message struct {
			Type    string          `json:"type"`
			ID      string          `json:"id,omitempty"`
			Payload json.RawMessage `json:"payload,omitempty"`
		}

		if err := json.Unmarshal(msg, &message); err != nil {
			log.Printf("[WS] Error parsing message: %v", err)
			sendErrorMessage(conn, "", "Invalid message format")
			continue
		}

		// Handle message based on type
		switch message.Type {
		case "connection_init":
			// Connection initialization
			log.Printf("[WS] Connection initialized")
			sendMessage(conn, "connection_ack", "", nil)

		case "start":
			// Start subscription
			var payload struct {
				Query         string                 `json:"query"`
				Variables     map[string]interface{} `json:"variables,omitempty"`
				OperationName string                 `json:"operationName,omitempty"`
			}

			if err := json.Unmarshal(message.Payload, &payload); err != nil {
				log.Printf("[WS] Error parsing subscription payload: %v", err)
				sendErrorMessage(conn, message.ID, "Invalid subscription payload")
				continue
			}

			log.Printf("[WS] Starting subscription %s: %s", message.ID, payload.Query)

			// Create quit channel for this subscription
			quit := make(chan struct{})
			subscriptions[message.ID] = quit

			// Start subscription goroutine
			go func(id string, query string, variables map[string]interface{}, quit <-chan struct{}) {
				// This is a simplified version - in a real implementation,
				// you would use resolver's subscription resolvers to stream events

				// For demonstration, we'll simulate delivery updates every few seconds
				ticker := time.NewTicker(5 * time.Second)
				defer ticker.Stop()

				statuses := []string{"PACKED", "OUT_FOR_DELIVERY", "DELIVERED"}
				statusIndex := 0

				for {
					select {
					case <-ticker.C:
						// Generate a sample delivery update
						data := map[string]interface{}{
							"data": map[string]interface{}{
								"deliveryUpdated": map[string]interface{}{
									"id":        fmt.Sprintf("%d", rand.Intn(1000)),
									"timestamp": time.Now().Format(time.RFC3339),
									"status":    statuses[statusIndex],
									"purchase": map[string]interface{}{
										"id":       variables["purchaseId"],
										"bankTxId": "TX" + fmt.Sprintf("%d", rand.Intn(10000)),
									},
								},
							},
						}

						// Send the update to the client
						sendMessage(conn, "data", id, data)

						// Cycle through statuses
						statusIndex = (statusIndex + 1) % len(statuses)

					case <-quit:
						log.Printf("[WS] Subscription %s terminated", id)
						return
					}
				}
			}(message.ID, payload.Query, payload.Variables, quit)

		case "stop":
			// Stop subscription
			if quit, ok := subscriptions[message.ID]; ok {
				close(quit)
				delete(subscriptions, message.ID)
				log.Printf("[WS] Stopped subscription %s", message.ID)
			}
			sendMessage(conn, "complete", message.ID, nil)

		case "connection_terminate":
			// Connection termination requested by client
			log.Printf("[WS] Connection termination requested")
			return

		default:
			log.Printf("[WS] Unknown message type: %s", message.Type)
		}
	}
}

// sendMessage sends a message to the WebSocket client
func sendMessage(conn *websocket.Conn, messageType, id string, payload interface{}) {
	msg := map[string]interface{}{
		"type": messageType,
	}

	if id != "" {
		msg["id"] = id
	}

	if payload != nil {
		msg["payload"] = payload
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("[WS] Error sending message: %v", err)
	}
}

// sendErrorMessage sends an error message to the WebSocket client
func sendErrorMessage(conn *websocket.Conn, id string, errorMessage string) {
	msg := map[string]interface{}{
		"type": "error",
		"payload": map[string]interface{}{
			"message": errorMessage,
		},
	}

	if id != "" {
		msg["id"] = id
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("[WS] Error sending error message: %v", err)
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// corsMiddleware adds CORS headers to responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization")

		// Handle OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Log the incoming request
		log.Printf("[HTTP] %s %s", r.Method, r.URL.Path)

		// Process the request
		next.ServeHTTP(w, r)
	})
}

// Update the Playground handler to support subscriptions
func playgroundHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`
<!DOCTYPE html>
<html>
<head>
    <meta charset=utf-8/>
    <meta name="viewport" content="user-scalable=no, initial-scale=1.0, minimum-scale=1.0, maximum-scale=1.0, minimal-ui">
    <title>GraphQL Playground</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/graphql-playground-react@1.7.22/build/static/css/index.css" />
    <link rel="shortcut icon" href="https://cdn.jsdelivr.net/npm/graphql-playground-react@1.7.22/build/favicon.png" />
    <script src="https://cdn.jsdelivr.net/npm/graphql-playground-react@1.7.22/build/static/js/middleware.js"></script>
</head>
<body>
    <div id="root">
        <style>
            body {
                background-color: rgb(23, 42, 58);
                font-family: Open Sans, sans-serif;
                height: 90vh;
            }
            #root {
                height: 100%;
                width: 100%;
                display: flex;
                align-items: center;
                justify-content: center;
            }
            .loading {
                font-size: 32px;
                font-weight: 200;
                color: rgba(255, 255, 255, .6);
                margin-left: 20px;
            }
            img {
                width: 78px;
                height: 78px;
            }
            .title {
                font-weight: 400;
            }
        </style>
        <img src='https://cdn.jsdelivr.net/npm/graphql-playground-react@1.7.22/build/logo.png' alt=''>
        <div class="loading">
            Loading <span class="title">GraphQL Playground</span>...
        </div>
    </div>
    <script>window.addEventListener('load', function (event) {
      const root = document.getElementById('root');
      root.classList.add('playgroundIn');
      const httpEndpoint = window.location.origin + '/graphql';
      const wsEndpoint = window.location.origin.replace('http', 'ws') + '/graphql/ws';
      GraphQLPlayground.init(root, { 
        endpoint: httpEndpoint,
        subscriptionEndpoint: wsEndpoint
      });
    })</script>
</body>
</html>
`))
}
