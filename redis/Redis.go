package redis

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/stanley2058/shorturl/structure"
)

var ctx = context.Background()
var redisConnection *redis.Client

func GetConnection() *redis.Client {
	if redisConnection == nil {
		db, err := strconv.Atoi(os.Getenv("REDIS_DB"))
		if err != nil {
			// fallback to default db
			db = 0
		}
		redisConnection = redis.NewClient(&redis.Options{
			Addr:     os.Getenv("REDIS_URL"),
			Username: os.Getenv("REDIS_USERNAME"),
			Password: os.Getenv("REDIS_PASSWORD"),
			DB:       db, // use default DB
		})
	}
	return redisConnection
}

func Save(key string, value string) error {
	return GetConnection().Set(ctx, key, value, 0).Err()
}

func Get(key string) (string, error) {
	return GetConnection().Get(ctx, key).Result()
}

func GetAllKeys() ([]string, error) {
	return GetConnection().Keys(ctx, "*").Result()
}

func GetAllEntries() (map[string]string, error) {
	keys, err := GetAllKeys()
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	for _, key := range keys {
		value, err := Get(key)
		if err != nil {
			return nil, err
		}
		result[key] = value
	}
	return result, nil
}

func Delete(key string) error {
	return GetConnection().Del(ctx, key).Err()
}

// Generate a random key consisting of {length} characters [0-9a-zA-Z], and make sure it's not already in use
//
// https://stackoverflow.com/questions/22892120/how-to-generate-a-random-string-of-a-fixed-length-in-go
func GenerateRandomKey(length ...int) string {
	var keyLength int
	if len(length) == 0 {
		keyLength = 6
	} else {
		keyLength = length[0]
	}
	var key string
	for {
		key = randSeq(keyLength)
		if val, err := Get(key); err != nil {
			break
		} else {
			var body structure.UrlObject
			err := json.Unmarshal([]byte(val), &body)
			if err != nil || !body.Activated {
				break
			}
		}
	}
	return key
}

var letters = []rune("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
