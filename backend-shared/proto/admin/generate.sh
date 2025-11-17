#!/bin/bash

# Generate Go code from proto files
# Run this script from the backend-shared directory

echo "Generating Go code from proto files..."

# Install required tools if not present
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/admin/admin.proto

echo "Proto generation complete!"

