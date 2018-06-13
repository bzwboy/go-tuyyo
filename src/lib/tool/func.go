package tool

import (
	"time"
	"encoding/json"
	"fmt"
	"strconv"
	"runtime"
	"os"
	"strings"
	"bytes"
	"lib/ltlog"
)

// 获取当前时间戳
func GetCurrTimeStamp() string {
	return string(time.Now().Unix())
}

// 解析json结构
// jsonStr json字符串字节流
// v 结构体类型
func DecodeJson(jsonStr []byte, v interface{}) (err error) {
	err = json.Unmarshal(jsonStr, v)
	if err != nil {
		fmt.Println("error:", err)
		return err
	}

	return nil
}

func GetIntDate() (num int64) {
	t := time.Now()
	str := fmt.Sprintf("%d%02d%d%02d%02d", t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute())
	dateInt, _ := strconv.Atoi(str)
	num = int64(dateInt)
	return
}

func GetLongDateString() (result string) {
	t := time.Now()
	result = fmt.Sprintf("%d-%02d-%d %02d:%02d", t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute())
	return
}

// debug convenions
func Debug(position string) string {
	if _, file, line, ok := runtime.Caller(2); ok {
		return fmt.Sprintf("["+position+"]", file, line)
	}
	return "!!NotContent!!"
}

/*
获取文件长度大小
*/
func GetFileLength(fileName string) int64 {
	fileInfo, err := os.Stat(fileName)
	if err != nil {
		ltlog.Println(ltlog.Lerror, err)
	}

	return fileInfo.Size()
}

/**
写文件，方便调试
 */
func PutContent(fileName, content string) int {
	_, file, line, _ := runtime.Caller(1)
	fileArr := strings.Split(file, "/")
	var start = -1
	for i, name := range fileArr {
		if name == "src" {
			start = i + 1
			break
		}
	}
	runFile := strings.Join(fileArr[start:], "/")

	dateString := GetLongDateString()
	prefix := fmt.Sprintf("%s %s(%d) ", dateString, runFile, line)
	byteContent := []byte(prefix + content + "\n")

	fh, err := os.Create(fileName)
	if err != nil {
		ltlog.Println(ltlog.Lerror, err)
	}

	n, err := fh.WriteAt(byteContent, GetFileLength(fileName))
	if err != nil {
		ltlog.Println(ltlog.Lerror, err)
	}

	return n
}

/*
连接字符串
 */
func JoinString(str ...string) string {
	var buf bytes.Buffer
	for _, s := range str {
		buf.WriteString(s)
	}

	return buf.String()
}
