package pubsub

import (
	"time"
)

type Interface interface {
	PublishLastUsed(idRoute string, lastUsed time.Time)
	SubscribeLastUsed(idRoute string, updateLastUsed func(id string, lastUsed time.Time) error)
	PublishIsRunning(idRoute string, isRunning bool)
	SubscribeIsRunning(idRoute string, updateIsRunning func(id string, isRunning bool) error)
	Unsubscribe(idRoute string)
}
