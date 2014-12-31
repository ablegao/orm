package orm

import (
	"database/sql"
	"fmt"
	"strings"
)

var NULL_LIMIT = [2]int{0, 0}
var databases = map[string]*Database{}

type Database struct {
	*sql.DB
	name           string
	driverName     string
	dataSourceName string
}

func (self *Database) Conn() (err error) {
	self.DB, err = sql.Open(self.driverName, self.dataSourceName)
	return
}

func NewDatabase(name, driverName, dataSourceName string) (database *Database, err error) {
	if database, ok := databases[name]; !ok {
		database = new(Database)
		database.name = name
		database.driverName = driverName
		database.dataSourceName = dataSourceName
		databases[name] = database
		err = database.Conn()
	} else {
		err = database.Ping()
	}
	return
}

type ParmaField struct {
	name string
	val  interface{}
}

/**
 传参解析
**/
type Params struct {
	connname  string
	stmt      *sql.Stmt
	tbname    string
	where     []ParmaField
	or        []ParmaField
	set       []ParmaField
	fields    []string
	order     []string
	limit     [2]int
	insertsql string
}

func (self *Params) GetWhereLen() int {
	return len(self.where)
}
func (self *Params) GetOrLen() int {
	return len(self.or)
}
func (self *Params) GetSetLen() int {
	return len(self.set)
}

func (self *Params) init() {
	if len(self.connname) == 0 {
		self.connname = "default"
	}
}

func (self *Params) SetTable(tbname string) {
	self.tbname = tbname

}

func (self *Params) SetField(fields ...string) {
	self.fields = fields
}

func (self *Params) Filter(name string, val interface{}) {
	self.where = append(self.where, ParmaField{name, val})
}
func (self *Params) FilterOr(name string, val interface{}) {
	self.or = append(self.or, ParmaField{name, val})
}

// 添加修改
func (self *Params) SetChange(name string, val interface{}) {
	self.set = append(self.set, ParmaField{name, val})
}

func (self Params) All() (rows *sql.Rows, err error) {
	//rows, err = self.db.Query(self.execSelect())
	//	self.stmt, err = self.db.Prepare()
	if db, ok := databases[self.connname]; !ok {
		panic("Database " + self.connname + " not defined.")
		return
	} else {

		sql, val := driversql[self.connname](self).Select()
		rows, err = db.Query(sql, val...)
	}

	return
}

func (self Params) One() (row *sql.Row) {
	//rows, err = self.db.Query(self.execSelect())
	//	self.stmt, err = self.db.Prepare()
	if db, ok := databases[self.connname]; ok {

		sql, val := driversql[self.connname](self).Select()
		row = db.QueryRow(sql, val...)
	}
	return
}
func (self Params) Delete() (res sql.Result, err error) {

	if db, ok := databases[self.connname]; ok {

		sql, val := driversql[self.connname](self).Delete()
		self.stmt, err = db.Prepare(sql)
		res, err = self.stmt.Exec(val...)

	} else {
		panic("Database " + self.connname + " not defined.")
	}
	return
}
func (self Params) Count() (int64, error) {
	if db, ok := databases[self.connname]; ok {
		sql, val := driversql[self.connname](self).Count()
		row := db.QueryRow(sql, val...)

		var c int64
		if err := row.Scan(&c); err == nil {
			return c, nil
		} else {
			return 0, err
		}
	} else {
		panic("Database " + self.connname + " not defined.")
	}

	return 0, nil
}

func (self Params) Save() (bool, int64, error) {
	db, ok := databases[self.connname]
	if !ok {
		panic("Database " + self.connname + " not defined.")
	}
	if c, _ := self.Count(); c > 0 {
		sql, val := driversql[self.connname](self).Update()
		self.stmt, _ = db.Prepare(sql)
		res, err := self.stmt.Exec(val...)

		if err != nil {
			return false, 0, err
		}
		a, b := res.RowsAffected()
		return false, a, b
	} else {
		sql, val := driversql[self.connname](self).Insert()
		self.stmt, _ = db.Prepare(sql)
		res, err := self.stmt.Exec(val...)
		if err != nil {
			return true, 0, err
		}
		a, b := res.LastInsertId()
		return true, a, b
	}

}

func (self *Params) GetTableName() string {
	tbname := ""
	if tb := strings.Split(self.tbname, "."); len(tb) > 1 {
		tbname = fmt.Sprintf("`%s`.`%s`", tb[0], tb[1])
	} else {
		tbname = "`" + self.tbname + "`"
	}
	return tbname
}

/*where

where 条件:
__exact        精确等于 like 'aaa'
 __iexact    精确等于 忽略大小写 ilike 'aaa'
 __contains    包含 like '%aaa%'
 __icontains    包含 忽略大小写 ilike '%aaa%'，但是对于sqlite来说，contains的作用效果等同于icontains。
__gt    大于
__gte    大于等于
__ne    不等于
__lt    小于
__lte    小于等于
__in     存在于一个list范围内
__startswith   以...开头
__istartswith   以...开头 忽略大小写
__endswith     以...结尾
__iendswith    以...结尾，忽略大小写
__range    在...范围内
__year       日期字段的年份
__month    日期字段的月份
__day        日期字段的日
__isnull=True/False


**/
func (self Params) _w(a string) string {
	typ := ""
	if bb := strings.Split(a, "__"); len(bb) > 1 {
		a = bb[0]
		typ = bb[1]
	}
	patten := ""
	switch typ {
	case "gt":
		patten = "`%s`>?"
	case "gte":
		patten = "`%s`>=?"
	case "lt":
		patten = "`%s`<?"
	case "lte":
		patten = "`%s`<=?"
	case "ne":
		patten = "`%s`<>?"
	case "add":
		return fmt.Sprintf("`%s`=`%s`+?", a, a)
	case "sub":
		return fmt.Sprintf("`%s`=`%s`-?", a, a)
	case "mult":
		return fmt.Sprintf("`%s`=`%s`*?", a, a)
	case "div":
		return fmt.Sprintf("`%s`=`%s`/?", a, a)
	case "asc":
		patten = "`%s` ASC"
	case "desc":
		patten = "`%s` DESC"
	default:
		patten = "`%s`=?"
	}
	return fmt.Sprintf(patten, a)
}
