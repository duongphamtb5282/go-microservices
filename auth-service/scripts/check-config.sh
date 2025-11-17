#!/bin/bash
cd "$(dirname "$0")"

echo "========================================="
echo "Config Check"
echo "========================================="
echo ""

echo "1. Checking config.yaml file:"
echo "---"
grep -A 3 "^authorization:" config/config.yaml
echo "---"
echo ""

echo "2. Checking if auth-service binary exists:"
ls -lh auth-service 2>/dev/null || echo "Binary not found"
echo ""

echo "3. Starting service with config debugging:"
echo "---"
export DATABASE_USERNAME=auth_user
export DATABASE_PASSWORD=auth_password

# Kill existing
lsof -ti:8085 | xargs kill -9 2>/dev/null
sleep 1

# Start and capture first 30 lines
timeout 5 ./auth-service 2>&1 | head -50 | grep -i "authorization\|mode\|config\|loading" || echo "No config logs found"
echo "---"
echo ""

# Kill it
lsof -ti:8085 | xargs kill -9 2>/dev/null

echo "4. Testing direct config parsing:"
cat > /tmp/test-config.go << 'EOF'
package main

import (
	"fmt"
	"github.com/spf13/viper"
)

func main() {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")
	
	if err := v.ReadInConfig(); err != nil {
		fmt.Printf("Error reading config: %v\n", err)
		return
	}
	
	mode := v.GetString("authorization.mode")
	enabled := v.GetBool("authorization.enabled")
	
	fmt.Printf("Authorization mode: %s\n", mode)
	fmt.Printf("Authorization enabled: %v\n", enabled)
}
EOF

cd "$(dirname "$0")"
go run /tmp/test-config.go
rm /tmp/test-config.go

echo ""
echo "========================================="

