// Copyright 2014 The 开心易点 All rights reserved.
// 非开源，公司内部程序， 不可转送带走，不可未经公司允许发布到公网、个人网盘、电子邮箱。不可做非公司电脑或非公司存储设备的备份。
// 不可对外泄露程序内部逻辑，程序结构，不可泄露程序数据结构以及相关加密算法。

// dbapi web接口

package main

import (
	"net/http"

	"github.com/ablegao/dbapi"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	http.Handle("/db", new(MyHandler))

}

type ErrorMess struct {
	Error bool
	Code  int
}

type SelectMess struct {
	ErrorMess
	Number int
	Rows   []map[string]interface{}
}

type MyHandler struct {
}

func (self *MyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	params, err := dbapi.NewDbParams("mysql", "happy:Cyup3EdezxW@tcp(192.168.0.50)/m2048?charset=utf8")
	if err != nil {

	}
	params.Unmarshal(r.Form)
	params.Method = r.Method

	switch params.RequestType {
	case "Select":

	}

	//w.Write([]byte())
}

func main() {
	http.ListenAndServe(":8801", nil)
}
