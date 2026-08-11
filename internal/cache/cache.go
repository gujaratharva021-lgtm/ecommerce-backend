package cache

import (
"context"
"encoding/json"
"time"

"github.com/gujaratharva021-lgtm/ecommerce-backend/internal/database"
)

// Get retrieves a cached value and unmarshals it into dest.
// Returns (found, error). found=false and error=nil means a clean cache miss.
func Get(ctx context.Context, key string, dest interface{}) (bool, error) {
if !database.RedisEnabled {
return false, nil
}
val, err := database.RDB.Get(ctx, key).Result()
if err != nil {
// redis.Nil (key not found) and any other error are both treated
// as a miss - callers should just fall through to the DB.
return false, nil
}
if err := json.Unmarshal([]byte(val), dest); err != nil {
return false, err
}
return true, nil
}

// Set stores a value in the cache with the given TTL. Silently does nothing
// if Redis is disabled.
func Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
if !database.RedisEnabled {
return nil
}
data, err := json.Marshal(value)
if err != nil {
return err
}
return database.RDB.Set(ctx, key, data, ttl).Err()
}

// DeleteByPrefix removes all keys matching a prefix (e.g. "products:list:*").
// Used for cache invalidation when underlying data changes.
func DeleteByPrefix(ctx context.Context, prefix string) error {
if !database.RedisEnabled {
return nil
}
iter := database.RDB.Scan(ctx, 0, prefix+"*", 100).Iterator()
var keys []string
for iter.Next(ctx) {
keys = append(keys, iter.Val())
}
if err := iter.Err(); err != nil {
return err
}
if len(keys) > 0 {
return database.RDB.Del(ctx, keys...).Err()
}
return nil
}

// Delete removes a single exact key.
func Delete(ctx context.Context, key string) error {
if !database.RedisEnabled {
return nil
}
return database.RDB.Del(ctx, key).Err()
}
