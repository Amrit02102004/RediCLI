package windows

import (
    "context"
    "fmt"
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