package utils

import (
    "context"
    "fmt"
    "strings"
    "time"
    "github.com/redis/go-redis/v9"
)

type RedisConnection struct {
    client *redis.Client
    ctx    context.Context
}

func NewRedisConnection() *RedisConnection {
    return &RedisConnection{
        ctx: context.Background(),
    }
}

func (rc *RedisConnection) Connect(host string, port string) error {
    rc.client = redis.NewClient(&redis.Options{
        Addr:     fmt.Sprintf("%s:%s", host, port),
        Password: "", // no password set
        DB:       0,  // use default DB
    })

    // Test the connection
    _, err := rc.client.Ping(rc.ctx).Result()
    if err != nil {
        return fmt.Errorf("failed to connect to Redis: %v", err)
    }

    return nil
}

func (rc *RedisConnection) Close() error {
    if rc.client != nil {
        return rc.client.Close()
    }
    return nil
}

func (rc *RedisConnection) IsConnected() bool {
    return rc.client != nil
}

func (rc *RedisConnection) GetAllKeys() ([]string, error) {
    if rc.client == nil {
        return nil, fmt.Errorf("not connected to Redis")
    }
    
    return rc.client.Keys(rc.ctx, "*").Result()
}

func (rc *RedisConnection) GetValue(key string) (string, error) {
    if rc.client == nil {
        return "", fmt.Errorf("not connected to Redis")
    }
    
    return rc.client.Get(rc.ctx, key).Result()
}

func (rc *RedisConnection) GetTTL(key string) (time.Duration, error) {
    if rc.client == nil {
        return 0, fmt.Errorf("not connected to Redis")
    }
    
    return rc.client.TTL(rc.ctx, key).Result()
}

func (rc *RedisConnection) ExecuteCommand(cmd string) (interface{}, error) {
    if rc.client == nil {
        return nil, fmt.Errorf("not connected to Redis")
    }
    
    // Split the command into parts
    parts := strings.Fields(cmd)
    if len(parts) == 0 {
        return nil, fmt.Errorf("empty command")
    }
    
    // Create a slice of interfaces starting with the command
    args := make([]interface{}, len(parts))
    for i, part := range parts {
        args[i] = part
    }
    
    // Execute the command
    return rc.client.Do(rc.ctx, args...).Result()
}

func (rc *RedisConnection) SetKeyWithTTL(key string, value string, ttl time.Duration) error {
    if rc.client == nil {
        return fmt.Errorf("not connected to Redis")
    }

    // If TTL is 0, set the key without expiration
    if ttl == 0 {
        return rc.client.Set(rc.ctx, key, value, 0).Err()
    }

    // Set key with specified TTL
    return rc.client.Set(rc.ctx, key, value, ttl).Err()
}

func (rc *RedisConnection) UpdateKey(key string, value string, keepTTL bool) error {
    if rc.client == nil {
        return fmt.Errorf("not connected to Redis")
    }

    // If keepTTL is true, we'll first get the current TTL
    var currentTTL time.Duration
    var err error
    if keepTTL {
        currentTTL, err = rc.client.TTL(rc.ctx, key).Result()
        if err != nil {
            return fmt.Errorf("error checking TTL: %v", err)
        }
    }

    // Set the new value
    if keepTTL && currentTTL > 0 {
        // Set with the existing TTL
        return rc.client.Set(rc.ctx, key, value, currentTTL).Err()
    } else {
        // Set without TTL
        return rc.client.Set(rc.ctx, key, value, 0).Err()
    }
}

func (rc *RedisConnection) KeyExists(key string) (bool, error) {
    if rc.client == nil {
        return false, fmt.Errorf("not connected to Redis")
    }

    // Check if the key exists
    exists, err := rc.client.Exists(rc.ctx, key).Result()
    if err != nil {
        return false, err
    }

    return exists == 1, nil
}

// Optional: Refresh data method if needed
func (rc *RedisConnection) RefreshData() ([]string, error) {
    if rc.client == nil {
        return nil, fmt.Errorf("not connected to Redis")
    }

    // Get all keys
    keys, err := rc.GetAllKeys()
    if err != nil {
        return nil, err
    }

    // Optionally, you can add more complex data retrieval logic here
    return keys, nil
}