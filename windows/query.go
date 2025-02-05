package windows

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/Amrit02102004/RediCLI/utils"
)

type QueryCondition struct {
	TTLOperator    string // >, <, ==
	TTLValue       int    // milliseconds
	ValuePattern   string // SQL LIKE pattern or regex pattern
	KeyPattern     string // SQL LIKE pattern or regex pattern
	IsRegexValue   bool   // true if value pattern is regex
	IsRegexKey     bool   // true if key pattern is regex
	ConnectionName string
}

// ParseQuery parses a query string into QueryCondition
func ParseQuery(query string) (*QueryCondition, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	condition := &QueryCondition{}

	// Extract connection name
	fromIndex := strings.Index(query, "from")
	if fromIndex == -1 {
		return nil, fmt.Errorf("missing 'from' clause")
	}
	// selectPart := query[:fromIndex]
	remainingQuery := query[fromIndex:]

	// Parse connection name
	parts := strings.Fields(remainingQuery)
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid query format")
	}
	condition.ConnectionName = parts[1]

	// Parse WHERE conditions
	whereIndex := strings.Index(remainingQuery, "where")
	if whereIndex == -1 {
		return condition, nil // No conditions specified
	}

	conditions := remainingQuery[whereIndex+5:]
	condParts := strings.Split(conditions, "and")

	for _, part := range condParts {
		part = strings.TrimSpace(part)

		// Parse TTL condition
		if strings.Contains(part, "ttl") {
			if err := parseTTLCondition(part, condition); err != nil {
				return nil, err
			}
			continue
		}

		// Parse value pattern
		if strings.Contains(part, "value") {
			if err := parseValueCondition(part, condition); err != nil {
				return nil, err
			}
			continue
		}

		// Parse key pattern
		if strings.Contains(part, "key") {
			if err := parseKeyCondition(part, condition); err != nil {
				return nil, err
			}
			continue
		}
	}

	return condition, nil
}

func parseTTLCondition(condition string, qc *QueryCondition) error {
	parts := strings.Fields(condition)
	if len(parts) < 3 {
		return fmt.Errorf("invalid TTL condition format")
	}

	// Extract operator
	var operator string
	switch parts[1] {
	case ">", "<", "==":
		operator = parts[1]
	default:
		return fmt.Errorf("invalid TTL operator: %s", parts[1])
	}

	// Extract value
	value, err := strconv.Atoi(parts[2])
	if err != nil {
		return fmt.Errorf("invalid TTL value: %s", parts[2])
	}

	qc.TTLOperator = operator
	qc.TTLValue = value
	return nil
}

func parseValueCondition(condition string, qc *QueryCondition) error {
	if strings.Contains(condition, "like") {
		pattern := extractPattern(condition, "like")
		qc.ValuePattern = sqlLikeToRegex(pattern)
		qc.IsRegexValue = false
	} else if strings.Contains(condition, "regex") {
		pattern := extractPattern(condition, "regex")
		qc.ValuePattern = pattern
		qc.IsRegexValue = true
	} else {
		return fmt.Errorf("invalid value condition format")
	}
	return nil
}

func parseKeyCondition(condition string, qc *QueryCondition) error {
	if strings.Contains(condition, "like") {
		pattern := extractPattern(condition, "like")
		qc.KeyPattern = sqlLikeToRegex(pattern)
		qc.IsRegexKey = false
	} else if strings.Contains(condition, "regex") {
		pattern := extractPattern(condition, "regex")
		qc.KeyPattern = pattern
		qc.IsRegexKey = true
	} else {
		return fmt.Errorf("invalid key condition format")
	}
	return nil
}

func extractPattern(condition, keyword string) string {
	parts := strings.Split(condition, keyword)
	if len(parts) != 2 {
		return ""
	}
	pattern := strings.TrimSpace(parts[1])
	// Remove quotes if present
	pattern = strings.Trim(pattern, "'\"")
	return pattern
}

func sqlLikeToRegex(pattern string) string {
	// Escape special regex characters
	pattern = regexp.QuoteMeta(pattern)
	// Convert SQL LIKE wildcards to regex patterns
	pattern = strings.Replace(pattern, "%", ".*", -1)
	pattern = strings.Replace(pattern, "_", ".", -1)
	return "^" + pattern + "$"
}

// ExecuteQuery executes the query and returns matching keys and their values
func ExecuteQuery(redis *utils.RedisConnection, condition *QueryCondition) (map[string]string, error) {
	if !redis.IsConnected() {
		return nil, fmt.Errorf("not connected to Redis")
	}

	// Get all keys
	keys, err := redis.GetAllKeys()
	if err != nil {
		return nil, err
	}

	results := make(map[string]string)
	for _, key := range keys {
		// Check key pattern if specified
		if condition.KeyPattern != "" {
			matched, err := regexp.MatchString(condition.KeyPattern, key)
			if err != nil || !matched {
				continue
			}
		}

		// Check TTL if specified
		if condition.TTLOperator != "" {
			ttl, err := redis.GetTTL(key)
			if err != nil {
				continue
			}

			ttlMS := int(ttl / time.Millisecond)
			switch condition.TTLOperator {
			case ">":
				if ttlMS <= condition.TTLValue {
					continue
				}
			case "<":
				if ttlMS >= condition.TTLValue {
					continue
				}
			case "==":
				if ttlMS != condition.TTLValue {
					continue
				}
			}
		}

		// Check value pattern if specified
		if condition.ValuePattern != "" {
			value, err := redis.GetValue(key)
			if err != nil {
				continue
			}

			matched, err := regexp.MatchString(condition.ValuePattern, value)
			if err != nil || !matched {
				continue
			}

			results[key] = value
		} else {
			// If no value pattern specified, include the key-value pair
			value, err := redis.GetValue(key)
			if err != nil {
				continue
			}
			results[key] = value
		}
	}

	return results, nil
}