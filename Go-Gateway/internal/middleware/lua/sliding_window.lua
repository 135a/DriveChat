-- sliding_window.lua
-- 基于 ZSET 的滑动窗口限流算法（原子操作）
--
-- KEYS[1] = 限流键 (e.g., gateway:ratelimit:<ip>)
-- ARGV[1] = 窗口大小（毫秒）
-- ARGV[2] = 最大请求数
-- ARGV[3] = 当前时间戳（毫秒）
-- ARGV[4] = 唯一请求标识 (当前时间戳 + 随机值)
--
-- 返回: 0 = 允许, 1 = 拒绝

local key = KEYS[1]
local window = tonumber(ARGV[1])
local limit = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local member = ARGV[4]

-- 1. 清理过期窗口外的数据
redis.call('ZREMRANGEBYSCORE', key, 0, now - window)

-- 2. 统计当前窗口内的请求数
local count = redis.call('ZCARD', key)

-- 3. 判断是否超过限制
if count >= limit then
    return 1
end

-- 4. 记录本次请求
redis.call('ZADD', key, now, member)

-- 5. 设置 TTL，避免 ZSET 无限增长（窗口大小的 2 倍，单位秒）
local ttl = math.ceil(window / 1000) * 2
redis.call('EXPIRE', key, ttl)

return 0
