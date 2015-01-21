package orm

import (
	"testing"

	_ "github.com/go-sql-driver/mysql"
)

type userB struct {
	CacheModule
	Uid     int64  `field:"Id" index:"pk"  cache:"user" `
	Alias   string `field:"Alias"`
	Lingshi int64  `field:"Lingshi"	`
}

func (self *userB) GetTableName() string {
	return "user_disk"
}

func Test_connect(t *testing.T) {

	CacheConsistent.Add("127.0.0.1:6379")

	_, err := NewDatabase("default", "mysql", "happy:passwd@tcp(127.0.0.1:3306)/mydatabase?charset=utf8")
	if err != nil {
		t.Error(err)
	}
	b := new(userB)
	b.Uid = 10000
	b.Objects(b).Ca(b.Uid).One()

	if err != nil {
		t.Error(err)
	}
	t.Log(b.Alias, b.Lingshi)
	b.Incrby("Lingshi", 10)
	b.Save()
	t.Log(b.Lingshi)
}
