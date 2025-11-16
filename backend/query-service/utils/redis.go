package utils

import (
	"context"
	"docusage/query-service/config"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// 内存缓存实现
type MemoryCache struct {
	cache map[string]cacheItem
	mutex sync.RWMutex
}

type cacheItem struct {
	data       []byte
	expiration int64
}

var (
	memoryCache  *MemoryCache
	redisContext = context.Background()
)

// InitRedis 初始化内存缓存（替代Redis）
func InitRedis(cfg *config.Config, log *zap.Logger) error {
	log.Info("Initializing memory cache (Redis alternative)",
		zap.String("mode", "memory_cache"),
	)

	memoryCache = &MemoryCache{
		cache: make(map[string]cacheItem),
	}

	log.Info("Memory cache initialized successfully")
	return nil
}

// GetMemoryCache 获取内存缓存
func GetMemoryCache() *MemoryCache {
	return memoryCache
}

// Set 设置缓存
func (mc *MemoryCache) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	exp := int64(0)
	if expiration > 0 {
		exp = time.Now().Add(expiration).UnixNano()
	}

	mc.cache[key] = cacheItem{
		data:       value,
		expiration: exp,
	}
	return nil
}

// Get 获取缓存
func (mc *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	item, exists := mc.cache[key]
	if !exists {
		return nil, fmt.Errorf("key not found")
	}

	// 检查是否过期
	if item.expiration > 0 && time.Now().UnixNano() > item.expiration {
		return nil, fmt.Errorf("key expired")
	}

	return item.data, nil
}

// Del 删除缓存
func (mc *MemoryCache) Del(ctx context.Context, keys ...string) error {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	for _, key := range keys {
		delete(mc.cache, key)
	}
	return nil
}

// Keys 查找匹配模式的键
func (mc *MemoryCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	var keys []string
	// 简化实现，仅支持简单的通配符
	for key := range mc.cache {
		// 这里实现一个简单的模式匹配
		if pattern == "*" || (pattern[:len(pattern)-1] == key[:len(pattern)-1] && pattern[len(pattern)-1] == '*') {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

// CacheSearchResult 缓存搜索结果
func CacheSearchResult(ctx context.Context, key string, result interface{}, expiration time.Duration) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return memoryCache.Set(ctx, key, data, expiration)
}

// GetCachedSearchResult 获取缓存的搜索结果
func GetCachedSearchResult(ctx context.Context, key string, result interface{}) error {
	data, err := memoryCache.Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, result)
}

// InvalidateUserCache 使用户相关缓存失效
func InvalidateUserCache(ctx context.Context, userID uint) error {
	pattern := fmt.Sprintf("search:user:%d:*", userID)
	keys, err := memoryCache.Keys(ctx, pattern)
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return memoryCache.Del(ctx, keys...)
	}

	return nil
}

// RecordSearchHistory 记录搜索历史
func RecordSearchHistory(ctx context.Context, userID uint, queryText string, maxItems int) error {
	// 简化的历史记录实现
	historyKey := fmt.Sprintf("search:history:user:%d", userID)

	// 获取现有历史
	historyData, err := memoryCache.Get(ctx, historyKey)
	var history []string
	if err == nil {
		json.Unmarshal(historyData, &history)
	}

	// 检查并移除重复项
	newHistory := []string{queryText}
	for _, item := range history {
		if item != queryText && len(newHistory) < maxItems {
			newHistory = append(newHistory, item)
		}
	}

	// 限制大小
	if len(newHistory) > maxItems {
		newHistory = newHistory[:maxItems]
	}

	// 保存回缓存
	historyBytes, _ := json.Marshal(newHistory)
	return memoryCache.Set(ctx, historyKey, historyBytes, 30*24*time.Hour)
}

// GetSearchHistory 获取搜索历史
func GetSearchHistory(ctx context.Context, userID uint, limit int) ([]string, error) {
	historyKey := fmt.Sprintf("search:history:user:%d", userID)
	historyData, err := memoryCache.Get(ctx, historyKey)
	if err != nil {
		return []string{}, nil
	}

	var history []string
	json.Unmarshal(historyData, &history)

	// 限制返回数量
	if len(history) > limit {
		history = history[:limit]
	}

	return history, nil
}

// CacheUserInfo 缓存用户信息
func CacheUserInfo(ctx context.Context, userID uint, userInfo interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("user:info:%d", userID)
	data, err := json.Marshal(userInfo)
	if err != nil {
		return err
	}

	return memoryCache.Set(ctx, key, data, expiration)
}

// GetCachedUserInfo 获取缓存的用户信息
func GetCachedUserInfo(ctx context.Context, userID uint, userInfo interface{}) error {
	key := fmt.Sprintf("user:info:%d", userID)
	data, err := memoryCache.Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, userInfo)
}

// RateLimit 限流函数
func RateLimit(ctx context.Context, key string, limit int, duration time.Duration) (bool, error) {
	// 简化的限流实现
	countKey := fmt.Sprintf("rate_limit:%s", key)

	// 获取当前计数
	data, err := memoryCache.Get(ctx, countKey)
	var count int64 = 0
	if err == nil {
		json.Unmarshal(data, &count)
	}

	// 增加计数
	count++

	// 如果是第一次访问，设置过期时间
	if count == 1 {
		countBytes, _ := json.Marshal(count)
		memoryCache.Set(ctx, countKey, countBytes, duration)
	} else {
		// 更新计数
		countBytes, _ := json.Marshal(count)
		memoryCache.Set(ctx, countKey, countBytes, 0)
	}

	// 检查是否超过限制
	return count <= int64(limit), nil
}

// CheckRedisHealth 检查内存缓存健康状态
func CheckRedisHealth() bool {
	return memoryCache != nil
}

// CloseRedis 关闭Redis连接
// CloseRedis 关闭内存缓存
func CloseRedis() error {
	if memoryCache == nil {
		return nil
	}

	// 清空缓存并重置
	memoryCache.mutex.Lock()
	defer memoryCache.mutex.Unlock()
	memoryCache.cache = make(map[string]cacheItem)

	return nil
}

// GetRedisStats 获取内存缓存统计信息
func GetRedisStats() map[string]interface{} {
	stats := make(map[string]interface{})
	stats["mode"] = "memory_cache"
	stats["status"] = "connected"

	if memoryCache != nil {
		memoryCache.mutex.RLock()
		stats["key_count"] = len(memoryCache.cache)
		memoryCache.mutex.RUnlock()
	} else {
		stats["key_count"] = 0
	}

	return stats
}
