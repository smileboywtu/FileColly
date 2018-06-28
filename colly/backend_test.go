// Test Suit for backend
package colly

import (
	"testing"
	"github.com/go-redis/redis"
)

var opts = &redis.Options{
	Addr:     "127.0.0.1:6379",
	DB:       0,
	Password: "",
}
var DestQueueName = "cache:dest"

func TestNewRedisWriter(t *testing.T) {
	inst, errs := NewRedisWriter(opts, DestQueueName, 500)
	if errs != nil {
		t.Fatal(errs.Error())
	}

	if inst.Check() {
		t.Log("test create new backend inst done")
	} else {
		t.Error("test create new backend inst error")
	}
}