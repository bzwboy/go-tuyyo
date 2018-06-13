package conf

import (
	"fmt"
	"strconv"
	"errors"
)

var ErrNil = errors.New("conf: nil returned")

// convert to string
func String(reply interface{}, err error) (string, error) {
	if err != nil {
		return "", err
	}

	switch reply := reply.(type) {
	case []byte:
		return string(reply), nil
	case string:
		return reply, nil
	case nil:
		return "", ErrNil
	}
	return "", fmt.Errorf("conf: unexpected type for String, got type %T", reply)
}

// convert to int
func Int(reply interface{}, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	switch reply := reply.(type) {
	case int64:
		x := int(reply)
		if int64(x) != reply {
			return 0, strconv.ErrRange
		}
		return x, nil
	case []byte:
		n, err := strconv.ParseInt(string(reply), 10, 0)
		return int(n), err
	case nil:
		return 0, ErrNil
	}
	return 0, fmt.Errorf("conf: unexpected type for Int, got type %T", reply)
}

// convert to bool
func Bool(reply interface{}, err error) (bool, error) {
	if err != nil {
		return false, err
	}
	switch reply := reply.(type) {
	case int64:
		return reply != 0, nil
	case []byte:
		return strconv.ParseBool(string(reply))
	case nil:
		return false, ErrNil
	}
	return false, fmt.Errorf("conf: unexpected type for Bool, got type %T", reply)
}

// convert to int64
func Int64(reply interface{}, err error) (int64, error) {
	if err != nil {
		return 0, err
	}
	switch reply := reply.(type) {
	case int64:
		return reply, nil
	case []byte:
		n, err := strconv.ParseInt(string(reply), 10, 64)
		return n, err
	case nil:
		return 0, ErrNil
	}
	return 0, fmt.Errorf("conf: unexpected type for Int64, got type %T", reply)
}
