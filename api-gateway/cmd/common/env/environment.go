package env

import (
	"fmt"
	"github.com/slink-go/api-gateway/cmd/common/str"
	"github.com/slink-go/logging"
	"github.com/xhit/go-str2duration/v2"
	"os"
	"strconv"
	"strings"
	"time"
)

func String(key string) (string, error) {
	result := os.Getenv(key)
	if result == "" {
		return "", fmt.Errorf("%s must be set", key)
	}
	return result, nil
}
func StringWithPrefix(key, prefix string) (string, error) {
	result := os.Getenv(key)
	if result == "" {
		return "", fmt.Errorf("%s must be set", key)
	}
	if !strings.HasPrefix(result, prefix) {
		return "", fmt.Errorf("%s must have the prefix \"%s\".", key, prefix)
	}
	return result, nil
}
func StringOrDefault(key string, defaultValue string) string {
	variable := os.Getenv(key)
	if variable == "" {
		return defaultValue
	}
	return variable
}
func BoolOrDefault(key string, defaultValue bool) bool {
	variable := os.Getenv(key)
	if variable == "" {
		return defaultValue
	}
	result, err := strconv.ParseBool(variable)
	if err != nil {
		return false
	}
	return result
}
func Int64OrDefault(key string, defaultValue int64) int64 {
	variable := os.Getenv(key)
	if variable == "" {
		return defaultValue
	}
	result, err := strconv.ParseInt(variable, 10, 64)
	if err != nil {
		return -1
	}
	return result
}
func DurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	variable := os.Getenv(key)
	if variable == "" {
		return defaultValue // default timeout
	} else {
		v, err := str2duration.ParseDuration(variable)
		if err != nil {
			logging.GetLogger("str").Warning("could not parse duration config value %s: %s", variable, err.Error())
			return defaultValue
		}
		return v
	}
}
func StringArrayOrEmpty(key string) []string {
	variable := os.Getenv(key)
	if variable != "" {
		return str.Slice(variable, ",")
	} else {
		return make([]string, 0)
	}
}
