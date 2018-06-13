package main

import (
	"testing"
	lt "lib/longtooth"
	"time"
	"strconv"
	"fmt"
)

func Benchmark_send(b *testing.B)  {
	ltId := "1.1.2.3830.2548.4042"

	for i := 0; i < b.N; i++ {
		content := strconv.Itoa(int(time.Now().UnixNano()))
		sendId := strconv.Itoa(i)
		sendTime := fmt.Sprint(time.Now().Unix())

		sd := &lt.Communication{
			CommProperty: lt.CommProperty{
				LtId:     ltId,
				Data:     content,
				SendId:   sendId,
				SendTime: sendTime,
			},
			OnSucc: sendSuccV1,
			OnFail: sendFailV1,
		}

		go lt.Send(sd)
	}
}

func sendFailV1(cmp *lt.CommProperty, st *lt.RespStatus) {
	fmt.Println(cmp.SendId)
}

func sendSuccV1(cmp *lt.CommProperty, st *lt.RespStatus) {
	fmt.Println(cmp.SendId)
}
