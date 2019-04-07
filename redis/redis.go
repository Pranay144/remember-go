package redis

import (
	"bytes"
	"context"
	"encoding/gob"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rocketlaunchr/remember-go"
)

type RedisStore struct {
	pool *redis.Pool
}

func NewRedisStore(redisPool *redis.Pool) *RedisStore {
	return &RedisStore{
		pool: redisPool,
	}
}

func (c *RedisStore) Conn(ctx context.Context) (remember.Cacher, error) {

	conn, err := c.pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}

	return &RedisConn{
		conn: conn,
	}, nil
}

type RedisConn struct {
	conn redis.Conn
}

func (c *RedisConn) StorePointer() bool {
	return true
}

func (c *RedisConn) Get(key string) (interface{}, bool, error) {

	val, err := redis.Bytes(c.conn.Do("GET", key))
	if err != nil {
		if err == redis.ErrNil {
			// Key not found
			return nil, false, nil
		}
		return nil, false, err
	}

	var output interface{}

	err = gob.NewDecoder(bytes.NewBuffer(val)).Decode(&output)
	if err != nil {
		return nil, true, err // Could not decode cached data
	}

	return output, true, nil
}

func (c *RedisConn) Set(key string, expiration time.Duration, itemToStore interface{}) error {

	// Convert item to bytes
	b := new(bytes.Buffer)
	err := gob.NewEncoder(b).Encode(itemToStore)
	if err != nil {
		return err
	}

	_, err = c.conn.Do("SET", key, b.Bytes(), "EX", int(expiration.Seconds()))
	return err
}

func (c *RedisConn) Close() {
	c.conn.Close()
}
