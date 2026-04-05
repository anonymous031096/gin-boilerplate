package middleware

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"gin-boilerplate/configs"

	"github.com/redis/go-redis/v9"
)

const revokeChannel = "token:revoke"

type revokeEntry struct {
	Timestamp int64
	ExpiresAt time.Time
}

type revokeMessage struct {
	Key       string `json:"key"`
	Timestamp int64  `json:"ts"`
}

var revokeStore = struct {
	sync.RWMutex
	items map[string]revokeEntry
}{items: make(map[string]revokeEntry)}

// InitRevokeSubscriber starts Pub/Sub listener and cleanup goroutine.
// Call once at app startup.
func InitRevokeSubscriber(redisClient *redis.Client) {
	// Load existing revoke keys from Redis on startup
	loadFromRedis(redisClient)

	// Subscribe to Pub/Sub for real-time updates
	go func() {
		sub := redisClient.Subscribe(context.Background(), revokeChannel)
		for msg := range sub.Channel() {
			var rm revokeMessage
			if err := json.Unmarshal([]byte(msg.Payload), &rm); err != nil {
				continue
			}
			storeSet(rm.Key, rm.Timestamp)
		}
	}()

	// Cleanup expired entries every minute
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			storeCleanup()
		}
	}()
}

// IsRevoked checks in-memory store. No Redis call.
func IsRevoked(userID string, deviceID string, iat int64) bool {
	revokeStore.RLock()
	defer revokeStore.RUnlock()

	now := time.Now()

	if e, ok := revokeStore.items[revokeKey(userID, deviceID)]; ok {
		if now.Before(e.ExpiresAt) && iat < e.Timestamp {
			return true
		}
	}

	if e, ok := revokeStore.items[revokeAllKey(userID)]; ok {
		if now.Before(e.ExpiresAt) && iat < e.Timestamp {
			return true
		}
	}

	return false
}

// publishRevoke saves to Redis + publishes to Pub/Sub + saves to local memory.
func publishRevoke(redisClient *redis.Client, key string) {
	cfg := configs.Get()
	now := time.Now().Unix()

	// Redis (persistence + backup for new instances)
	redisClient.Set(context.Background(), key, now, cfg.JWT.RefreshTTL)

	// Pub/Sub (sync other instances)
	msg, _ := json.Marshal(revokeMessage{Key: key, Timestamp: now})
	redisClient.Publish(context.Background(), revokeChannel, msg)

	// Local memory (this instance, immediate)
	storeSet(key, now)
}

func storeSet(key string, timestamp int64) {
	ttl := configs.Get().JWT.RefreshTTL
	revokeStore.Lock()
	revokeStore.items[key] = revokeEntry{
		Timestamp: timestamp,
		ExpiresAt: time.Now().Add(ttl),
	}
	revokeStore.Unlock()
}

func storeCleanup() {
	revokeStore.Lock()
	now := time.Now()
	for k, v := range revokeStore.items {
		if now.After(v.ExpiresAt) {
			delete(revokeStore.items, k)
		}
	}
	revokeStore.Unlock()
}

// loadFromRedis loads existing revoke keys on startup (cold start recovery).
func loadFromRedis(redisClient *redis.Client) {
	ctx := context.Background()
	iter := redisClient.Scan(ctx, 0, "revoke:*", 0).Iterator()
	for iter.Next(ctx) {
		key := iter.Val()
		ts, err := redisClient.Get(ctx, key).Int64()
		if err == nil {
			storeSet(key, ts)
		}
	}
}
