package pubsub

import (
	"time"
)

type Interface interface {
	Publish(id string, lastUsed time.Time)
	Subscribe(id string, updateLastUsed func(id string, lastUsed time.Time) error)
	Unsubscribe(id string)
}
