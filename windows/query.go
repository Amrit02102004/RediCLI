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

type UpdateType int

const (
	UpdateValue UpdateType = iota
	UpdateKey
	UpdateTTL
)

type UpdateQuery struct {
	ConnectionName string
	UpdateType    UpdateType
	NewValue      string
	Condition     *QueryCondition
}

// ParseUpdateQuery parses an update query string
func ParseUpdateQuery(query string) (*UpdateQuery, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	if !strings.HasPrefix(query, "update") {
		return nil, fmt.Errorf("query must start with 'update'")
	}

	updateQuery := &UpdateQuery{}

	// Split into main parts: UPDATE connection SET type = value WHERE conditions
	parts := strings.Split(query, "set")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid update query format, missing SET clause")
	}

	// Parse connection name
	connectionParts := strings.Fields(parts[0])
	if len(connectionParts) < 2 {
		return nil, fmt.Errorf("missing connection name")
	}
	updateQuery.ConnectionName = connectionParts[1]

	// Split SET and WHERE clauses
	setAndWhere := strings.Split(parts[1], "where")
	if len(setAndWhere) != 2 {
		return nil, fmt.Errorf("missing WHERE clause")
	}

	// Parse SET clause
	setClause := strings.TrimSpace(setAndWhere[0])
	if err := parseSetClause(setClause, updateQuery); err != nil {
		return nil, err
	}

	// Parse WHERE clause using existing QueryCondition parser
	whereClause := strings.TrimSpace(setAndWhere[1])
	condition, err := ParseQuery("select from " + updateQuery.ConnectionName + " where " + whereClause)
	if err != nil {
		return nil, fmt.Errorf("error parsing WHERE clause: %v", err)
	}
	updateQuery.Condition = condition

	return updateQuery, nil
}

func parseSetClause(setClause string, updateQuery *UpdateQuery) error {
	setParts := strings.Split(setClause, "=")
	if len(setParts) != 2 {
		return fmt.Errorf("invalid SET clause format")
	}

	updateType := strings.TrimSpace(setParts[0])
	newValue := strings.TrimSpace(setParts[1])
	newValue = strings.Trim(newValue, "'\"") // Remove quotes if present

	switch updateType {
	case "value":
		updateQuery.UpdateType = UpdateValue
	case "key":
		updateQuery.UpdateType = UpdateKey
	case "ttl":
		updateQuery.UpdateType = UpdateTTL
	default:
		return fmt.Errorf("invalid update type: %s", updateType)
	}

	updateQuery.NewValue = newValue
	return nil
}

// ExecuteUpdateQuery executes the update query
func ExecuteUpdateQuery(redis *utils.RedisConnection, query *UpdateQuery) (int, error) {
	if !redis.IsConnected() {
		return 0, fmt.Errorf("not connected to Redis")
	}

	// First get all matching keys based on the condition
	matches, err := ExecuteQuery(redis, query.Condition)
	if err != nil {
		return 0, err
	}

	updatedCount := 0

	// Process each matching key
	for key := range matches {
		switch query.UpdateType {
		case UpdateValue:
			err = updateValue(redis, key, query.NewValue)
		case UpdateKey:
			err = updateKey(redis, key, query.NewValue)
		case UpdateTTL:
			err = updateTTL(redis, key, query.NewValue)
		}

		if err != nil {
			continue // Skip to next key if there's an error
		}
		updatedCount++
	}

	return updatedCount, nil
}

func updateValue(redis *utils.RedisConnection, key, newValue string) error {
	// Get current TTL
	ttl, err := redis.GetTTL(key)
	if err != nil {
		return err
	}

	// Update value while preserving TTL
	return redis.SetKeyWithTTL(key, newValue, ttl)
}

func updateKey(redis *utils.RedisConnection, oldKey, newKey string) error {
	// Get current value
	value, err := redis.GetValue(oldKey)
	if err != nil {
		return err
	}

	// Get current TTL
	ttl, err := redis.GetTTL(oldKey)
	if err != nil {
		return err
	}

	// Set new key with same value and TTL
	err = redis.SetKeyWithTTL(newKey, value, ttl)
	if err != nil {
		return err
	}

	// Delete old key
	_, err = redis.ExecuteCommand(fmt.Sprintf("del %s", oldKey))
	return err
}

func updateTTL(redis *utils.RedisConnection, key, ttlStr string) error {
	// Parse TTL value (in milliseconds)
	ttlMs, err := strconv.ParseInt(ttlStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid TTL value: %v", err)
	}

	// Convert to duration
	ttl := time.Duration(ttlMs) * time.Millisecond

	// Get current value
	value, err := redis.GetValue(key)
	if err != nil {
		return err
	}

	// Update TTL
	return redis.SetKeyWithTTL(key, value, ttl)
}


type DeleteQuery struct {
	ConnectionName string
	Condition     *QueryCondition
}

// ParseDeleteQuery parses a delete query string
func ParseDeleteQuery(query string) (*DeleteQuery, error) {
	query = strings.TrimSpace(strings.ToLower(query))
	if !strings.HasPrefix(query, "del from") {
		return nil, fmt.Errorf("query must start with 'del from'")
	}

	deleteQuery := &DeleteQuery{}

	// Split into main parts: DEL FROM connection WHERE conditions
	parts := strings.Split(query, "where")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid delete query format, missing WHERE clause")
	}

	// Parse connection name
	connectionParts := strings.Fields(parts[0])
	if len(connectionParts) < 3 {
		return nil, fmt.Errorf("missing connection name")
	}
	deleteQuery.ConnectionName = connectionParts[2]

	// Parse WHERE clause using existing QueryCondition parser
	whereClause := strings.TrimSpace(parts[1])
	condition, err := ParseQuery("select from " + deleteQuery.ConnectionName + " where " + whereClause)
	if err != nil {
		return nil, fmt.Errorf("error parsing WHERE clause: %v", err)
	}
	deleteQuery.Condition = condition

	return deleteQuery, nil
}

// ExecuteDeleteQuery executes the delete query and returns a confirmation function along with matched keys
func ExecuteDeleteQuery(redis *utils.RedisConnection, deleteQuery *DeleteQuery) (confirmFunc func() (int, error), matchedKeys []string, err error) {
    if !redis.IsConnected() {
        return nil, nil, fmt.Errorf("not connected to Redis")
    }

    // First get all matching keys based on the condition
    matches, err := ExecuteQuery(redis, deleteQuery.Condition)
    if err != nil {
        return nil, nil, err
    }

    // Get list of keys that would be deleted
    matchedKeys = make([]string, 0, len(matches))
    for key := range matches {
        matchedKeys = append(matchedKeys, key)
    }

    // Return confirmation function
    confirmFunc = func() (int, error) {
        deletedCount := 0
        for _, key := range matchedKeys {
            _, err := redis.ExecuteCommand(fmt.Sprintf("del %s", key))
            if err != nil {
                continue
            }
            deletedCount++
        }
        return deletedCount, nil
    }

    return confirmFunc, matchedKeys, nil
}