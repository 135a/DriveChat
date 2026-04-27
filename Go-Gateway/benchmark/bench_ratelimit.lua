-- bench_ratelimit.lua
-- wrk Lua 脚本：模拟不同 IP 的限流压测场景
--
-- 使用方法:
--   wrk -t4 -c200 -d30s -s benchmark/bench_ratelimit.lua http://localhost:8080/api/test
--

-- 随机 IP 池大小
local ip_pool_size = 50
local counter = 0

-- 生成随机 IP 地址
function random_ip()
    return string.format("%d.%d.%d.%d",
        math.random(1, 254),
        math.random(1, 254),
        math.random(1, 254),
        math.random(1, 254))
end

-- 预生成 IP 池
local ips = {}
for i = 1, ip_pool_size do
    ips[i] = random_ip()
end

-- 每次请求使用不同的 X-Forwarded-For 头来模拟不同客户端
request = function()
    counter = counter + 1
    local ip = ips[(counter % ip_pool_size) + 1]

    local headers = {}
    headers["X-Forwarded-For"] = ip
    headers["Content-Type"] = "application/json"

    return wrk.format("GET", nil, headers)
end

-- 统计响应状态码
local status_200 = 0
local status_429 = 0
local status_403 = 0
local status_other = 0

response = function(status, headers, body)
    if status == 200 then
        status_200 = status_200 + 1
    elseif status == 429 then
        status_429 = status_429 + 1
    elseif status == 403 then
        status_403 = status_403 + 1
    else
        status_other = status_other + 1
    end
end

done = function(summary, latency, requests)
    local total = status_200 + status_429 + status_403 + status_other
    io.write("\n========= Rate Limit Benchmark Results =========\n")
    io.write(string.format("Total Requests:     %d\n", total))
    io.write(string.format("  200 OK:           %d (%.1f%%)\n", status_200, status_200/total*100))
    io.write(string.format("  429 Rate Limited: %d (%.1f%%)\n", status_429, status_429/total*100))
    io.write(string.format("  403 Blacklisted:  %d (%.1f%%)\n", status_403, status_403/total*100))
    io.write(string.format("  Other:            %d (%.1f%%)\n", status_other, status_other/total*100))
    io.write("==================================================\n")
end
