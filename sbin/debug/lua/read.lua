-- redis-cli --eval select.lua
local ck, er
local nr = {}

ck = "meeting:queue"
nr[1] = "--> ["..ck.."] <--"
nr[2], er = redis.call("LRANGE", ck, 0, -1)

ck = "msg:user"
nr[3] = "--> ["..ck.."] <--"
nr[4], er = redis.call("SMEMBERS", ck)

ck = "msg:1:1.1.2.3830.2548.4042"
nr[5] = "--> ["..ck.."] <--"
nr[6], er = redis.call("ZRANGE", ck, 0, -1, "WITHSCORES")

return nr

