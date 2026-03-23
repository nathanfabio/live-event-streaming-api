package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
)



type Client struct {
	*redis.Client
}

func NewClient(addr string) (*Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &Client{client}, nil
}