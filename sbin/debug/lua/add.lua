-- redis-cli --eval add.lua
local function assemble(gid, sid)
    local msg_tpl, err

    msg_tpl = '{"group_id":"<gid>","lt_id":"1.1.2.3830.2548.4042","send_id":"<sid>","data":"{\\"status\\":{\\"code\\":0,\\"msg\\":\\"\\"},\\"cmd\\":\\"creategroup\\",\\"send_id\\":\\"<sid>\\"}"}'
    msg_tpl, err = string.gsub(msg_tpl, "<gid>", gid)
    msg_tpl, err = string.gsub(msg_tpl, "<sid>", sid)
    return msg_tpl
end

local queue_key = "meeting:queue"
local nr, er, msg
local max = 20

redis.call("FLUSHALL")
for i=1,max,1 do
    msg = assemble(1,i)
    redis.call("RPUSH", queue_key, msg)
end

-- nr, er = redis.call("LRANGE", queue_key, 0, -1)
-- return nr
