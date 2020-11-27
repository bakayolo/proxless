package redis

import (
	"fmt"
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
		m:      make(map[string]*redis.PubSub),
	}
}

func (r *RedisClient) PublishLastUsed(idRoute string, lastUsed time.Time) {
	idChannel := genLastUsedChannelName(idRoute)
	err := r.client.Publish(idChannel, lastUsed.Unix()).Err()
	if err != nil {
		logger.Errorf(err, "Cannot PUBLISH message to Redis channel %s", idChannel)
	}
}

func (r *RedisClient) SubscribeLastUsed(idRoute string, updateLastUsed func(id string, lastUsed time.Time) error) {
	idChannel := genLastUsedChannelName(idRoute)
	if _, ok := r.m[idChannel]; !ok {
		r.m[idChannel] = r.client.Subscribe(idChannel)

		go func() {
			for {
				msg, ok := <-r.m[idChannel].Channel()

				if !ok {
					logger.Debugf("Could not receive message from channel %s - might have been closed", idChannel)
					return
				}

				timestampInSec, err := strconv.Atoi(msg.Payload)

				if err != nil {
					logger.Errorf(err, "Could not unmarshal payload %s from channel %s", msg.Payload, idChannel)
					continue
				}

				err = updateLastUsed(idRoute, time.Unix(int64(timestampInSec), 0))
				if err != nil {
					logger.Errorf(err, "Could not update lastUsed in route id %s", idChannel)
				}
			}
		}()
	}
}

func (r *RedisClient) PublishIsRunning(idRoute string, isRunning bool) {
	idChannel := genIsRunningChannelName(idRoute)
	err := r.client.Publish(idChannel, isRunning).Err()
	if err != nil {
		logger.Errorf(err, "Cannot PUBLISH message to Redis channel %s", idChannel)
	}
}

func (r *RedisClient) SubscribeIsRunning(idRoute string, updateIsRunning func(id string, isRunning bool) error) {
	idChannel := genIsRunningChannelName(idRoute)
	if _, ok := r.m[idChannel]; !ok {
		r.m[idChannel] = r.client.Subscribe(idChannel)

		go func() {
			for {
				msg, ok := <-r.m[idChannel].Channel()

				if !ok {
					logger.Debugf("Could not receive message from channel %s - might have been closed", idChannel)
					return
				}

				isRunning, err := strconv.ParseBool(msg.Payload)

				if err != nil {
					logger.Errorf(err, "Could not unmarshal payload %s from channel %s", msg.Payload, idChannel)
					continue
				}

				err = updateIsRunning(idRoute, isRunning)
				if err != nil {
					logger.Errorf(err, "Could not update isRunning field in route id %s", idChannel)
				}
			}
		}()
	}
}

func (r *RedisClient) Unsubscribe(idRoute string) {
	// TODO commenting this because of the `nil` exception due to the redis library
	// can repro remotely but not locally - To Be Fixed later

	// idLastUsedChannel := genLastUsedChannelName(idRoute)
	// err := r.m[idLastUsedChannel].Close()
	//
	// if err != nil {
	// 	logger.Errorf(err, "Could not close the subscription to channel %s", idLastUsedChannel)
	// }

	// idIsRunningChannel := genIsRunningChannelName(idRoute)
	// err = r.m[idIsRunningChannel].Close()
	//
	// if err != nil {
	// 	logger.Errorf(err, "Could not close the subscription to channel %s", idIsRunningChannel)
	// }
}

func genLastUsedChannelName(id string) string {
	return fmt.Sprintf("last_used_%s", id)
}

func genIsRunningChannelName(id string) string {
	return fmt.Sprintf("is_running_%s", id)
}
