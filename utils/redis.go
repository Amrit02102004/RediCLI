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