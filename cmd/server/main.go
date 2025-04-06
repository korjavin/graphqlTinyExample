package main

import (
	"log"
	"net/http"
	"os"

	"github.com/graph-gophers/graphql-go/relay"
	_ "github.com/lib/pq"

	"github.com/korjavin/graphqlTinyExample/pkg/graphql"
	"github.com/korjavin/graphqlTinyExample/pkg/models"
	"github.com/korjavin/graphqlTinyExample/pkg/repository"
)

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

	// Set up HTTP handler
	http.Handle("/graphql", corsMiddleware(&relay.Handler{Schema: schema}))

	// Serve GraphQL Playground for interactive API exploration
	http.HandleFunc("/", playgroundHandler)

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("Server started at http://localhost:%s/", port)
	log.Printf("GraphQL endpoint: http://localhost:%s/graphql", port)
	log.Printf("GraphQL Playground: http://localhost:%s/", port)

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Failed to start server: %v", err)
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

// playgroundHandler serves the GraphQL Playground interface
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
      const endpoint = window.location.origin + '/graphql';
      GraphQLPlayground.init(root, { endpoint });
    })</script>
</body>
</html>
`))
}
