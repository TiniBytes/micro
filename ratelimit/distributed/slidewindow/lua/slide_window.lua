-- 滑动窗口限流
local key = KEYS[1]
local window = tonumber(ARGV[1])
local threshold = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local begin = now - window

-- 删除[-inf, min]区间的元素
redis.call('ZREMRANGEBYSCORE', key, '-inf', begin)
-- 计算所有区间的元素数量
local cnt = redis.call('ZCOUNT', key, begin, '+inf')

if cnt < threshold then
    -- score 和 member都设置now
    redis.call('ZADD', key, now, now)
    redis.call('PEXPIRE', key, window)
    return "true"
else
    -- 执行限流
    return "false"
end