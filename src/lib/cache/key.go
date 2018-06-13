package cache

var keys = map[string]string{
	"mainQueue": "meeting:queue", // LIST
	"userList":  "msg:user", // SET
	"runList":  "msg:user@run",
	"userMsg":   "msg:", // LIST
}

// 消息队列
func keyMsgQueue() (ck string, err error) {
	return keys["mainQueue"], nil
}

// 用户消息索引
func keyUserList() (ck string, err error) {
	return keys["userList"], nil
}

// 正在运行的用户
// 防止异常情况下丢失用户，例如：段错误
func keyRunList() (ck string, err error) {
	return keys["runList"], nil
}

// 用户消息
func keyUserMsg(user string) (ck string, err error) {
	return keys["userMsg"] + user, nil
}
