# Test Scripts

This folder contains comprehensive test scripts for the auth service and related components.

## 📋 Available Test Scripts

### **1. test-all-components.sh**

Master test script that runs all component tests with interactive menu.

**Usage:**

```bash
./scripts/test-all-components.sh
```

**Features:**

- Interactive menu to select which tests to run
- Prerequisites checking
- Quick health checks
- Comprehensive test summary

### **2. test-correlation-id.sh**

Tests Correlation ID functionality in HTTP requests.

**Usage:**

```bash
./scripts/test-correlation-id.sh
```

**Tests:**

- Correlation ID generation
- Custom correlation ID handling
- Multiple request correlation
- Different endpoint testing
- Logging verification
- Middleware testing

### **3. test-graphql.sh**

Tests GraphQL service functionality.

**Usage:**

```bash
./scripts/test-graphql.sh
```

**Tests:**

- GraphQL introspection
- Query execution
- Mutation testing
- Subscription testing
- Error handling
- Performance testing

### **4. test-kafka.sh**

Tests Kafka messaging functionality.

**Usage:**

```bash
./scripts/test-kafka.sh
```

**Tests:**

- Kafka broker connectivity
- Topic management
- Producer functionality
- Consumer functionality
- Auth service integration
- Performance testing
- Error handling

## 🚀 Quick Start

### **Run All Tests:**

```bash
./scripts/test-all-components.sh
```

### **Run Individual Tests:**

```bash
# Correlation ID tests
./scripts/test-correlation-id.sh

# GraphQL tests
./scripts/test-graphql.sh

# Kafka tests
./scripts/test-kafka.sh
```

## 📋 Prerequisites

### **Required Services:**

- **Auth Service**: Running on port 8085
- **GraphQL Service**: Running on port 8080 (optional)
- **Kafka Broker**: Running on port 9092 (optional)

### **Starting Services:**

```bash
# Auth Service
go run main.go

# GraphQL Service
cd ../graphql-service
go run main.go

# Kafka (Docker)
docker-compose up -d kafka
```

## 🔧 Test Configuration

### **Service URLs:**

- Auth Service: `http://localhost:8085`
- GraphQL Service: `http://localhost:8080`
- Kafka Broker: `localhost:9092`

### **Kafka Configuration:**

- Topic: `auth-events`
- Consumer Group: `auth-service-test`

## 📊 Test Results

Each test script provides:

- ✅ Success indicators
- ❌ Error messages
- 📋 Detailed summaries
- 💡 Next steps and recommendations

## 🐛 Troubleshooting

### **Common Issues:**

1. **Auth Service Not Running**

   ```bash
   go run main.go
   ```

2. **GraphQL Service Not Running**

   ```bash
   cd ../graphql-service
   go run main.go
   ```

3. **Kafka Not Accessible**

   ```bash
   docker-compose up -d kafka
   ```

4. **Permission Denied**
   ```bash
   chmod +x scripts/*.sh
   ```

## 📝 Notes

- All scripts are designed to be run from the auth-service root directory
- Scripts include comprehensive error handling and user feedback
- Test results are displayed in real-time
- Scripts can be run individually or as part of the master test suite
