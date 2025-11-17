package graphql

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"graphql-service/internal/interfaces/graphql/resolvers"

	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
)

// TestServer represents the GraphQL test server
type TestServer struct {
	router       *mux.Router
	userResolver *resolvers.UserResolver
	logger       interface{} // Replace with actual logger type
}

// NewTestServer creates a new GraphQL test server
func NewTestServer(userResolver *resolvers.UserResolver, logger interface{}) *TestServer {
	server := &TestServer{
		router:       mux.NewRouter(),
		userResolver: userResolver,
		logger:       logger,
	}

	// Setup routes
	server.setupRoutes()

	return server
}

// setupRoutes configures the HTTP routes
func (s *TestServer) setupRoutes() {
	// GraphQL playground
	s.router.Handle("/", playground.Handler("GraphQL Test Playground", "/query"))

	// GraphQL endpoint
	s.router.HandleFunc("/query", s.graphqlHandler)

	// Health check
	s.router.HandleFunc("/health", s.healthHandler)

	// Test endpoints
	s.router.HandleFunc("/test/users", s.testUsersHandler)
	s.router.HandleFunc("/test/create-user", s.testCreateUserHandler)
}

// graphqlHandler creates the GraphQL handler
func (s *TestServer) graphqlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Process GraphQL query
	response := s.processGraphQLQuery(req.Query, req.Variables)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// processGraphQLQuery processes GraphQL queries
func (s *TestServer) processGraphQLQuery(query string, variables map[string]interface{}) map[string]interface{} {
	// Simple GraphQL query processing
	// This is a simplified implementation for testing
	
	if query == "" {
		return map[string]interface{}{
			"data": map[string]interface{}{
				"message": "GraphQL test server is running",
				"timestamp": "2024-01-01T00:00:00Z",
			},
		}
	}

	// Handle introspection query
	if contains(query, "__schema") {
		return map[string]interface{}{
			"data": map[string]interface{}{
				"__schema": map[string]interface{}{
					"types": []map[string]interface{}{
						{"name": "User"},
						{"name": "Order"},
						{"name": "Product"},
						{"name": "Notification"},
					},
				},
			},
		}
	}

	// Handle user queries
	if contains(query, "users") {
		users, err := s.userResolver.Users(context.Background(), nil, nil)
		if err != nil {
			return map[string]interface{}{
				"errors": []map[string]interface{}{
					{"message": err.Error()},
				},
			}
		}

		return map[string]interface{}{
			"data": map[string]interface{}{
				"users": users,
			},
		}
	}

	// Handle create user mutation
	if contains(query, "createUser") {
		// Extract input from variables
		input, ok := variables["input"].(map[string]interface{})
		if !ok {
			return map[string]interface{}{
				"errors": []map[string]interface{}{
					{"message": "Invalid input"},
				},
			}
		}

		user, err := s.userResolver.CreateUser(context.Background(), input)
		if err != nil {
			return map[string]interface{}{
				"errors": []map[string]interface{}{
					{"message": err.Error()},
				},
			}
		}

		return map[string]interface{}{
			"data": map[string]interface{}{
				"createUser": user,
			},
		}
	}

	// Default response
	return map[string]interface{}{
		"data": map[string]interface{}{
			"message": "GraphQL query processed",
			"query": query,
		},
	}
}

// testUsersHandler tests user queries
func (s *TestServer) testUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := s.userResolver.Users(context.Background(), nil, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": users,
		"count": len(users),
	})
}

// testCreateUserHandler tests user creation
func (s *TestServer) testCreateUserHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var input map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	user, err := s.userResolver.CreateUser(context.Background(), input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user": user,
		"message": "User created successfully",
	})
}

// healthHandler handles health check requests
func (s *TestServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"status": "healthy",
		"service": "graphql-test-server",
		"database": "mongodb",
		"timestamp": "2024-01-01T00:00:00Z",
	}
	
	json.NewEncoder(w).Encode(response)
}

// Start starts the GraphQL test server
func (s *TestServer) Start(port string) error {
	log.Printf("Starting GraphQL test server on port %s", port)
	return http.ListenAndServe(":"+port, s.router)
}

// Helper function to check if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		   len(s) > len(substr) && contains(s[1:], substr)
}
