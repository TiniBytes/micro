-- 限流设置
local val = redis.call('get', KEYS[1])
local expiration = ARGV[1]
local allow = tonumber(ARGV[2])
if val == false then
    if allow < 1 then
        -- 执行限流
        return "false"
    else
        -- set user-service 1 px 100s
        redis.call('set', KEYS[1], 1, 'PX', expiration)
        return "true"
    end
elseif tonumber(val) < allow then
    -- 有限流对象，但是没到阈值
    redis.call('incr', KEYS[1])
    return "true"
else
    -- 执行限流
    return "false"
end