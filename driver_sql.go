package orm

import (
	"fmt"
	"strings"
)

type driversqlType func(param Params) ModuleToSql

var driversql = map[string]driversqlType{
	"mysql": func(param Params) ModuleToSql { return MysqlModeToSql{param} },
}

type ModuleToSql interface {
	Select() (sql string, val []interface{})
	Delete() (sql string, val []interface{})
	Update() (sql string, val []interface{})
	Insert() (sql string, val []interface{})
	Count() (sql string, val []interface{})
	Instance(Params)
}

type MysqlModeToSql struct {
	Params
}

func (self MysqlModeToSql) Instance(param Params) {
	self.Params = param
}

func (self MysqlModeToSql) _where() (sql string, val []interface{}) {
	whereLen := self.Params.GetWhereLen()
	orLen := self.Params.GetOrLen()

	where := make([]string, whereLen)
	or := make([]string, orLen)

	val = make([]interface{}, whereLen+orLen)

	i := 0
	for _, w := range self.where {

		//where = append(where, self.Params._w(w.name))
		where[i] = self.Params._w(w.name)
		val[i] = w.val
		i = i + 1
	}

	for _, w := range self.Params.or {
		or[i] = self.Params._w(w.name)
		val[i] = w.val
		i = i + 1
	}

	sql = ""
	switch {
	case whereLen > 0 && orLen > 0:
		sql = sql + " WHERE " + strings.Join(where, " AND ") + " OR " + strings.Join(or, " OR ")
	case whereLen > 0 && orLen == 0:
		sql = sql + " WHERE " + strings.Join(where, " AND ")
	case orLen > 0 && whereLen == 0:
		sql = sql + " WHERE " + strings.Join(or, " OR ")
	}
	return
}
func (self MysqlModeToSql) _set() (sql string, val []interface{}) {
	set := make([]string, len(self.Params.set))
	val = make([]interface{}, len(self.Params.set))
	for i, v := range self.Params.set {
		set[i] = self.Params._w(v.name)
		val[i] = v
	}
	sql = " SET " + strings.Join(set, ",")
	return
}
func (self MysqlModeToSql) Insert() (sql string, val []interface{}) {
	sql, val = self._set()
	sql = fmt.Sprintf("INSERT INTO  %s %s ",
		self.Params.GetTableName(),
		sql,
	)
	return
}
func (self MysqlModeToSql) Update() (sql string, val []interface{}) {
	sql, val = self._set()
	sql = fmt.Sprintf("UPDATE  %s %s ",
		self.Params.GetTableName(),
		sql,
	)
	s, v := self._where()
	sql = sql + s
	val = append(val, v...)
	return
}

func (self MysqlModeToSql) Delete() (sql string, val []interface{}) {
	sql, val = self._where()

	sql = fmt.Sprintf("DELETE FROM %s %s ",
		self.Params.GetTableName(),
		sql,
	)

	return
}
func (self MysqlModeToSql) Select() (sql string, val []interface{}) {

	sql, val = self._where()
	sql = fmt.Sprintf("SELECT `%s` FROM %s  %s",
		strings.Join(self.Params.fields, "`,`"),
		self.Params.GetTableName(),
		sql,
	)

	if len(self.order) > 0 {
		sql = sql + " ORDER BY "
		ret := make([]string, len(self.order))
		for id, v := range self.Params.order {
			ret[id] = self.Params._w(v)
		}
		sql = sql + strings.Join(ret, ",")
	}

	if self.Params.limit != NULL_LIMIT {
		sql = sql + fmt.Sprintf(" LIMIT %d , %d", self.limit[0], self.limit[1])
	}

	return
}
func (self MysqlModeToSql) Count() (sql string, val []interface{}) {
	sql, val = self._where()
	sql = fmt.Sprintf("SELECT COUNT(*) FROM %s  %s ",
		self.Params.GetTableName(),
		sql,
	)
	return
}
