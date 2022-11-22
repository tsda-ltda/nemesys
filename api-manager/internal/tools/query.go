package tools

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// DefaultQuery returns the query value of the context.
// If query value is empty, returns the default value.
func DefaultQuery(c *gin.Context, key string, defaultValue string) string {
	raw := c.Query(key)
	if len(raw) > 0 {
		return raw
	}
	return defaultValue
}

// IntQuery returns the int value of a query value.
// If query value is "", returns default value.
// If query value is invalid int returns an error.
func IntQuery(c *gin.Context, key string, defaultValue int) (int, error) {
	raw := c.Query(key)
	if len(raw) == 0 {
		return defaultValue, nil
	}
	return strconv.Atoi(raw)
}

// IntRangeQuery returns the int value of a query value.
// If query value is "", returns default value.
// If query value is invalid returns an error.
// If query value is bigger than max value, returns max value.
// If query value is less than min value, returns min value.
func IntRangeQuery(c *gin.Context, key string, defaultValue int, max int, min int) (int, error) {
	raw := c.Query(key)
	if len(raw) == 0 {
		return defaultValue, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue, err
	}

	if v > max {
		v = max
	} else if v < min {
		v = min
	}

	return v, nil
}

// IntMaxQuery returns the int value of a query value.
// If query value is "", returns default value.
// If query value is invalid returns an error.
// If query value is bigger than max value, returns max value.
func IntMaxQuery(c *gin.Context, key string, defaultValue int, max int) (int, error) {
	raw := c.Query(key)
	if len(raw) == 0 {
		return defaultValue, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue, err
	}

	if v > max {
		v = max
	}

	return v, nil
}

// IntMinQuery returns the int value of a query value.
// If query value is "", returns default value.
// If query value is invalid returns an error.
// If query value is less than min value, returns min value.
func IntMinQuery(c *gin.Context, key string, defaultValue int, min int) (int, error) {
	raw := c.Query(key)
	if len(raw) == 0 {
		return defaultValue, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue, err
	}

	if v < min {
		v = min
	}

	return v, nil
}
