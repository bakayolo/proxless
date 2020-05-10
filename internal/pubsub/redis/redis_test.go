package redis

import (
	"testing"
)

// TODO does not work because alicebob does not support PUBLISH -.-
func TestPubSub(t *testing.T) {
	// s, err := miniredis.Run()
	// if err != nil {
	// 	panic(err)
	// }
	// defer s.Close()
	//
	// redisCli := NewRedisPubSub(s.Addr())
	//
	// id := "dummyId"
	// now := time.Now()
	// messageReceived := false
	//
	// redisCli.Subscribe(id, func(id string, lastUsed time.Time) error{
	// 	assert.Equal(t, time.Now(), lastUsed)
	// 	messageReceived = true
	// 	return nil
	// })
	//
	// redisCli.Publish(id, now)
	//
	// time.Sleep(time.Second)
	//
	// assert.Equal(t, true, messageReceived)
}
