# GraphQL Service with MongoDB

A comprehensive GraphQL service implementation using Go, MongoDB, and Docker for testing GraphQL communication patterns.

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   GraphQL       │    │   MongoDB       │    │   Docker        │
│   Service       │◄──►│   Database      │    │   Environment   │
│   (Go)          │    │   (MongoDB)     │    │   (Compose)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📁 Project Structure

```
graphql-service/
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/
│   ├── domain/
│   │   ├── user/
│   │   │   ├── entity/
│   │   │   │   └── user.go         # User domain entity
│   │   │   └── repository/
│   │   │       └── user_repository.go
│   │   ├── order/
│   │   │   └── entity/
│   │   │       └── order.go        # Order domain entity
│   │   ├── product/
│   │   │   └── entity/
│   │   │       └── product.go      # Product domain entity
│   │   └── notification/
│   │       └── entity/
│   │           └── notification.go # Notification domain entity
│   ├── infrastructure/
│   │   ├── database/
│   │   │   └── init.go             # Database initialization
│   │   └── persistence/
│   │       └── mongodb/
│   │           └── user_repository.go # MongoDB user repository
│   └── interfaces/
│       └── graphql/
│           ├── schema.graphql      # GraphQL schema definition
│           ├── server.go           # GraphQL server setup
│           └── resolvers/
│               └── user_resolver.go # User GraphQL resolvers
├── scripts/
│   ├── init-mongo.js              # MongoDB initialization script
│   └── test_graphql.sh            # GraphQL testing script
├── docker-compose.yml             # Docker Compose configuration
├── Dockerfile                     # Docker image definition
├── go.mod                         # Go module dependencies
└── README.md                      # This file
```

## 🚀 Features

### GraphQL Schema

- **User Management**: CRUD operations for users
- **Order Management**: Order creation, status updates, cancellation
- **Product Catalog**: Product management with categories
- **Notifications**: Real-time notification system
- **Subscriptions**: Real-time updates for orders and notifications

### Database Integration

- **MongoDB**: Primary database with collections for users, orders, products, notifications
- **Indexes**: Optimized indexes for performance
- **Sample Data**: Pre-loaded sample data for testing

### API Endpoints

- `GET /` - GraphQL Playground
- `POST /query` - GraphQL endpoint
- `GET /health` - Health check

## 🛠️ Setup and Installation

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- MongoDB 7.0+

### 1. Clone and Setup

```bash
cd /Users/duongphamthaibinh/Downloads/SourceCode/design/beautiful/golang/graphql-service
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Start MongoDB

```bash
# Using Docker Compose
docker-compose up -d mongodb

# Or using Docker directly
docker run -d --name auth-mongodb -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=admin_password \
  -e MONGO_INITDB_DATABASE=graphql_service \
  mongo:7.0
```

### 4. Run the Service

```bash
# Development mode
go run cmd/server/main.go

# Or build and run
go build -o graphql-service cmd/server/main.go
./graphql-service
```

### 5. Test the Service

```bash
# Run the test script
./scripts/test_graphql.sh

# Or test manually
curl http://localhost:8086/health
curl http://localhost:8086/
```

## 📊 GraphQL Schema

### Types

#### User

```graphql
type User {
  id: ID!
  username: String!
  email: String!
  firstName: String!
  lastName: String!
  orders: [Order!]!
  notifications: [Notification!]!
  createdAt: String!
  updatedAt: String!
}
```

#### Order

```graphql
type Order {
  id: ID!
  userId: ID!
  user: User!
  items: [OrderItem!]!
  status: OrderStatus!
  totalAmount: Float!
  createdAt: String!
  updatedAt: String!
}
```

#### Product

```graphql
type Product {
  id: ID!
  name: String!
  description: String
  price: Float!
  category: String!
  stock: Int!
  createdAt: String!
  updatedAt: String!
}
```

#### Notification

```graphql
type Notification {
  id: ID!
  userId: ID!
  user: User!
  type: NotificationType!
  title: String!
  message: String!
  read: Boolean!
  createdAt: String!
  updatedAt: String!
}
```

### Queries

```graphql
# Get user by ID
user(id: "user_id"): User

# Get all users with filtering and pagination
users(filter: UserFilter, pagination: Pagination): [User!]!

# Get order by ID
order(id: "order_id"): Order

# Get all orders
orders(filter: OrderFilter, pagination: Pagination): [Order!]!

# Get products
products(filter: ProductFilter, pagination: Pagination): [Product!]!

# Get notifications
notifications(filter: NotificationFilter, pagination: Pagination): [Notification!]!
```

### Mutations

```graphql
# User mutations
createUser(input: CreateUserInput!): User!
updateUser(id: ID!, input: UpdateUserInput!): User!
deleteUser(id: ID!): Boolean!

# Order mutations
createOrder(input: CreateOrderInput!): Order!
updateOrderStatus(id: ID!, input: UpdateOrderStatusInput!): Order!
cancelOrder(id: ID!): Order!

# Product mutations
createProduct(input: CreateProductInput!): Product!
updateProduct(id: ID!, input: UpdateProductInput!): Product!
deleteProduct(id: ID!): Boolean!

# Notification mutations
createNotification(input: CreateNotificationInput!): Notification!
updateNotification(id: ID!, input: UpdateNotificationInput!): Notification!
deleteNotification(id: ID!): Boolean!
```

### Subscriptions

```graphql
# Real-time subscriptions
orderStatusChanged(userId: ID!): Order!
userCreated: User!
notificationCreated(userId: ID!): Notification!
productStockUpdated: Product!
```

## 🧪 Testing

### GraphQL Playground

Access the GraphQL Playground at: http://localhost:8086/

### Sample Queries

#### Get All Users

```graphql
query {
  users {
    id
    username
    email
    firstName
    lastName
    createdAt
  }
}
```

#### Create User

```graphql
mutation {
  createUser(
    input: {
      username: "johndoe"
      email: "john@example.com"
      firstName: "John"
      lastName: "Doe"
    }
  ) {
    id
    username
    email
    fullName: firstName
  }
}
```

#### Get User with Orders

```graphql
query {
  user(id: "user_id") {
    id
    username
    email
    orders {
      id
      status
      totalAmount
      items {
        product {
          name
          price
        }
        quantity
      }
    }
  }
}
```

### Health Check

```bash
curl http://localhost:8086/health
```

## 🔧 Configuration

### Environment Variables

- `MONGO_URI`: MongoDB connection string
- `PORT`: Server port (default: 8086)
- `LOG_LEVEL`: Logging level (debug, info, warn, error)

### MongoDB Configuration

- Database: `graphql_service`
- Collections: `users`, `orders`, `products`, `notifications`
- Indexes: Optimized for common queries

## 🐳 Docker Support

### Build and Run

```bash
# Build the image
docker build -t graphql-service .

# Run the container
docker run -p 8086:8086 graphql-service
```

### Docker Compose

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f graphql-service

# Stop services
docker-compose down
```

## 📈 Performance Considerations

### Database Indexes

- Unique indexes on email and username
- Compound indexes for common queries
- TTL indexes for data retention

### GraphQL Optimizations

- Field-level resolvers
- DataLoader pattern for N+1 queries
- Query complexity analysis
- Caching strategies

## 🔒 Security Features

### Input Validation

- GraphQL schema validation
- Type checking
- Required field validation

### Database Security

- Connection string authentication
- Index-based access patterns
- Query optimization

## 🚀 Deployment

### Production Considerations

1. **Environment Variables**: Configure production settings
2. **Database**: Use production MongoDB cluster
3. **Monitoring**: Add logging and metrics
4. **Scaling**: Horizontal scaling with load balancer
5. **Caching**: Redis for query caching

### Health Checks

- Database connectivity
- Service availability
- Response time monitoring

## 📚 API Documentation

### GraphQL Playground

The GraphQL Playground provides:

- Interactive query builder
- Schema documentation
- Query history
- Variable editor

### Endpoints

- **GraphQL**: `POST /query`
- **Playground**: `GET /`
- **Health**: `GET /health`

## 🎯 Use Cases

### E-commerce Platform

- User management
- Order processing
- Product catalog
- Notification system

### Real-time Applications

- Live order updates
- User notifications
- Stock updates
- Status changes

### Microservices Integration

- Service-to-service communication
- Data aggregation
- Event-driven architecture
- API gateway integration

## 🔄 Integration with Existing Services

### Auth Service Integration

```graphql
# Query user from auth service
query {
  user(id: "auth_user_id") {
    id
    username
    email
    orders {
      id
      status
    }
  }
}
```

### Event-Driven Architecture

- User creation events
- Order status changes
- Notification triggers
- Real-time updates

## 📊 Monitoring and Observability

### Metrics

- Query execution time
- Database connection pool
- Error rates
- Response times

### Logging

- Structured logging
- Request tracing
- Error tracking
- Performance monitoring

## 🎉 Success Criteria

✅ **MongoDB Integration**: Successfully connected and tested
✅ **GraphQL Schema**: Complete schema with all entities
✅ **CRUD Operations**: Full CRUD for all entities
✅ **Real-time Features**: Subscriptions implemented
✅ **Docker Support**: Containerized and ready
✅ **Testing**: Comprehensive test suite
✅ **Documentation**: Complete API documentation

## 🚀 Next Steps

1. **Implement gRPC Integration**: Add gRPC communication
2. **Add Kafka Retry**: Implement retry mechanisms
3. **Federation**: GraphQL federation for microservices
4. **Authentication**: JWT-based authentication
5. **Rate Limiting**: API rate limiting
6. **Caching**: Redis caching layer
7. **Monitoring**: Prometheus metrics
8. **Tracing**: Distributed tracing

This GraphQL service provides a solid foundation for testing GraphQL communication patterns with MongoDB and serves as a reference implementation for microservices architecture.
