package redis

import (
	"github.com/go-redis/redis/v7"
	"kube-proxless/internal/logger"
	"kube-proxless/internal/pubsub"
	"strconv"
	"time"
)

type RedisClient struct {
	client *redis.Client
	m      map[string]*redis.PubSub
}

func NewRedisPubSub(redisURL string) pubsub.Interface {
	client := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: "", // no password set - the data there don't need to be protected
		DB:       0,  // use default DB
	})

	_, err := client.Ping().Result()
	if err != nil {
		// we don't return error - the proxy must still work even pubsub not working
		// it will just not be full HA
		logger.Errorf(err, "Cannot PING Redis - please check if further errors and fix Redis connection if needed")
	} else {
		logger.Infof("Proxless connected to Redis on %s", redisURL)
	}

	return &RedisClient{
		client: client,
	}
}

func (r *RedisClient) Publish(id string, lastUsed time.Time) {
	err := r.client.Publish(id, lastUsed.Unix()).Err()
	if err != nil {
		logger.Errorf(err, "Cannot PUBLISH message to Redis channel %s", id)
	}
}

func (r *RedisClient) Subscribe(id string, updateLastUsed func(id string, lastUsed time.Time) error) {
	if _, ok := r.m[id]; !ok {
		r.m[id] = r.client.Subscribe(id)

		go func() {
			for {
				msg := <-r.m[id].Channel()

				timestampInSec, err := strconv.Atoi(msg.Payload)

				if err != nil {
					logger.Errorf(err, "Could not unmarshal payload %s from channel %s", msg.Payload, id)
					continue
				}

				err = updateLastUsed(id, time.Unix(int64(timestampInSec), 0))
				if err != nil {
					logger.Errorf(err, "Could not update lastUsed in route id %s", id)
				}
			}
		}()
	}
}

func (r *RedisClient) Unsubscribe(id string) {
	err := r.m[id].Close()

	if err != nil {
		logger.Errorf(err, "Could not close the subscription to channel %s", id)
	}
}
