package orm

import "testing"

type userB struct {
	CacheModule
	Id        int64  `field:"id" index:"pk" auto:"true" cache:"user" `
	Udid      string `field:"udid" index:"index" cache:"udid"`
	Username  string `field:"username"`
	Token     string `field:"token"`
	Face      string `field:"face"`
	Level     int64  `field:"level"`
	Exp       int64  `field:"exp"`
	Fpoint    int64  `field:"fb"`
	LastLogin string `field:"last_login"`
	WarNum    int64  `field:"warnum"`
	WinNum    int64  `field:"winnum"`
}

func (self *userB) GetTableName() string {
	return "user"
}

func Test_connect(t *testing.T) {
	b := new(userB)
	b.Id = 1
	b.Objects(b)

}
