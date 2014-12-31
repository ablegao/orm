package main

import (
	"testing"

	"github.com/ablegao/dbapi"
)

func Test_select(t *testing.T) {
	//client := http.Client{}
	params := dbapi.Params{}
	params.SetField("a", "b", "c")
	params.SetTable("abc", "def")

	params.Filter("field__eq", 1)
	params.Filter("bbb_neq", 2)

	t.Log(params.GetSelectUrl().Encode())
}
