package orm

import "testing"

func Test_execSelect(t *testing.T) {
	params := Params{}
	params.SetField("a", "b", "c")
	params.SetTable("abc.def")

	params.Filter("field__gt", 1)
	params.Filter("bbb__lt", 2)
	params.SetChange("a", 1)
	params.SetChange("b__sub", 1)
	params.SetChange("c__div", 1)
	//t.Log(params.execSelect())
	str, _ := driversql["mysql"](params).Select()
	t.Log(str)
	str, _ = driversql["mysql"](params).Delete()
	t.Log(str)
	str, _ = driversql["mysql"](params).Insert()
	t.Log(str)
	str, _ = driversql["mysql"](params).Update()
	t.Log(str)
	str, _ = driversql["mysql"](params).Count()
	t.Log(str)
}
