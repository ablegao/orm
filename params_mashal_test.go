package orm

import (
	"testing"
)

func Test_select(t *testing.T) {
	//client := http.Client{}
	params := Params{}
	params.SetField("a", "b", "c")
	params.SetTable("abc.def")

	params.Filter("field__gt", 1)
	params.Filter("bbb__lt", 2)

}
