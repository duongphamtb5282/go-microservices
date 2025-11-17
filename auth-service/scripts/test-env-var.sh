#!/bin/bash

echo "=== Testing Environment Variable for Authorization Mode ==="
echo ""

# Test 1: Without env var
echo "Test 1: No AUTHORIZATION_MODE set"
unset AUTHORIZATION_MODE
go run -C . - <<'EOF'
package main
import (
    "fmt"
    "github.com/spf13/viper"
    "strings"
)
func main() {
    v := viper.New()
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath("./config")
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.ReadInConfig()
    mode := v.GetString("authorization.mode")
    fmt.Printf("Mode from Viper: '%s'\n", mode)
}
EOF
echo ""

# Test 2: With env var
echo "Test 2: With AUTHORIZATION_MODE=pingam"
export AUTHORIZATION_MODE=pingam
go run -C . - <<'EOF'
package main
import (
    "fmt"
    "os"
    "github.com/spf13/viper"
    "strings"
)
func main() {
    fmt.Printf("Environment AUTHORIZATION_MODE='%s'\n", os.Getenv("AUTHORIZATION_MODE"))
    v := viper.New()
    v.SetConfigName("config")
    v.SetConfigType("yaml")
    v.AddConfigPath("./config")
    v.AutomaticEnv()
    v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
    v.ReadInConfig()
    mode := v.GetString("authorization.mode")
    fmt.Printf("Mode from Viper: '%s'\n", mode)
}
EOF
echo ""
echo "=== Test Complete ==="

