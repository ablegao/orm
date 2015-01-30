// model解析支持

package orm

import (
	"errors"
	"reflect"
	"strings"
	"sync"
	"time"
)

type Module interface {
	GetTableName() string
}

func Objects(mode Module) Object {

	typ := reflect.TypeOf(mode).Elem()
	vals := []string{}
	for i := 0; i < typ.NumField(); i++ {
		if field := typ.Field(i).Tag.Get("field"); len(field) > 0 {
			vals = append(vals, field)
		}
	}

	obj := Object{}
	obj.SetTable(mode.GetTableName())
	obj.mode = mode
	obj.Params.Init()
	obj.Params.SetField(vals...)
	return obj
}

type Object struct {
	sync.RWMutex
	Params
	mode Module
}

func (self *Object) Objects(mode Module) *Object {
	self.Lock()
	defer self.Unlock()
	self.SetTable(mode.GetTableName())
	self.Init()

	typ := reflect.TypeOf(mode).Elem()
	vals := []string{}
	for i := 0; i < typ.NumField(); i++ {
		if field := typ.Field(i).Tag.Get("field"); len(field) > 0 {
			vals = append(vals, field)
		}
	}
	self.SetField(vals...)
	self.mode = mode
	return self
}

//修改数据
// name 结构字段名称
// val 结构数据
func (self *Object) Change(name string, val interface{}) *Object {
	self.Lock()
	defer self.Unlock()
	typ := reflect.TypeOf(self.mode).Elem()
	fieldName := strings.Split(name, "__")
	if field, ok := typ.FieldByName(fieldName[0]); ok && len(field.Tag.Get("field")) > 0 {
		name := field.Tag.Get("field")
		if len(fieldName) > 1 {
			name = name + "__" + fieldName[1]
		}
		self.Params.Change(name, val)
	}
	return self
}

//条件筛选
// name 结构字段名称
// val 需要过滤的数据值
func (self *Object) Filter(name string, val interface{}) *Object {
	self.Lock()
	defer self.Unlock()
	typ := reflect.TypeOf(self.mode).Elem()
	fieldName := strings.Split(name, "__")
	if field, ok := typ.FieldByName(fieldName[0]); ok && len(field.Tag.Get("field")) > 0 {
		name := field.Tag.Get("field")
		if len(fieldName) > 1 {
			name = name + "__" + fieldName[1]
		}
		self.Params.Filter(name, val)
	}
	return self
}
func (self *Object) FilterOr(name string, val interface{}) *Object {
	self.Lock()
	defer self.Unlock()
	typ := reflect.TypeOf(self.mode).Elem()
	fieldName := strings.Split(name, "__")
	if field, ok := typ.FieldByName(fieldName[0]); ok && len(field.Tag.Get("field")) > 0 {
		name := field.Tag.Get("field")
		if len(fieldName) > 1 {
			name = name + "__" + fieldName[1]
		}
		self.Params.FilterOr(name, val)
	}
	return self
}

// Filter 的一次传入版本 ， 不建议使用 , 因为map 循序不可控
func (self *Object) Filters(filters map[string]interface{}) *Object {
	self.Lock()
	defer self.Unlock()
	for k, v := range filters {
		self.Filter(k, v)
	}
	return self
}

// Order by 排序 ，
// Field__asc Field__desc
func (self *Object) Orderby(names ...string) *Object {
	typ := reflect.TypeOf(self.mode).Elem()
	for i, name := range names {
		fieldName := strings.Split(name, "__")
		if field, ok := typ.FieldByName(fieldName[0]); ok && len(field.Tag.Get("field")) > 0 {
			if name = field.Tag.Get("field"); len(name) > 0 {
				name = name + "__" + fieldName[1]
				names[i] = name
			}
		}
	}
	self.Params.order = names
	return self
}

// 分页支持
func (self *Object) Limit(page, steq int) *Object {
	self.Lock()
	defer self.Unlock()
	self.Params.limit = [2]int{page, steq}
	return self
}

//选择数据库
func (self *Object) Db(name string) *Object {
	self.Params.Db(name)
	return self
}

// 计算数量
func (self *Object) Count() (int64, error) {
	self.RLock()
	defer self.RUnlock()
	return self.Params.Count()
}

//删除数据
func (self *Object) Delete() (int64, error) {
	self.Lock()
	defer self.Unlock()
	valus := reflect.ValueOf(self.mode).Elem()
	for i := 0; i < valus.NumField(); i++ {
		typ := valus.Type().Field(i)
		val := valus.Field(i)
		if typ.Tag.Get("index") == "pk" {
			switch val.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if val.Int() > 0 {
					self.Params.Filter(typ.Name, val.Int())
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if val.Uint() > 0 {
					self.Params.Filter(typ.Name, val.Uint())
				}
			case reflect.Float32, reflect.Float64:
				if val.Float() > 0.0 {
					self.Params.Filter(typ.Tag.Get("field"), val.Float())
				}
			case reflect.String:
				if len(val.String()) > 0 {
					self.Params.Filter(typ.Tag.Get("field"), val.String())
				}
			default:
				switch val.Interface().(type) {
				case time.Time:
					self.Params.Filter(typ.Tag.Get("field"), val.Interface())
				}
			}
		}
	}
	if len(self.Params.where) > 0 {
		res, err := self.Params.Delete()
		if err != nil {
			return 0, err
		}
		return res.RowsAffected()
	} else {
		return 0, errors.New("No Where")
	}

}

//更新活添加
func (self *Object) Save() (bool, int64, error) {
	self.Lock()
	defer self.Unlock()
	valus := reflect.ValueOf(self.mode).Elem()
	if len(self.Params.set) == 0 {
		for i := 0; i < valus.NumField(); i++ {
			typ := valus.Type().Field(i)
			val := valus.Field(i)
			if self.hasRow {
				if len(typ.Tag.Get("field")) > 0 && typ.Tag.Get("index") != "pk" {
					self.Params.Change(typ.Tag.Get("field"), val.Interface())
				}
			} else {
				if len(typ.Tag.Get("field")) > 0 {
					self.Params.Change(typ.Tag.Get("field"), val.Interface())
				}
			}
		}
	}

	self.autoWhere()
	isNew, id, err := self.Params.Save()
	return isNew, id, err
}

func (self *Object) autoWhere() {
	valus := reflect.ValueOf(self.mode).Elem()
	if len(self.Params.where) == 0 {
		for i := 0; i < valus.NumField(); i++ {
			typ := valus.Type().Field(i)
			val := valus.Field(i)
			if len(typ.Tag.Get("field")) > 0 && typ.Tag.Get("index") == "pk" {
				switch val.Kind() {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if val.Int() > 0 {
						self.Params.Filter(typ.Tag.Get("field"), val.Int())
					}
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					if val.Uint() > 0 {
						self.Params.Filter(typ.Tag.Get("field"), val.Uint())
					}
				case reflect.Float32, reflect.Float64:
					if val.Float() > 0.0 {
						self.Params.Filter(typ.Tag.Get("field"), val.Float())
					}
				case reflect.String:
					if len(val.String()) > 0 {
						self.Params.Filter(typ.Tag.Get("field"), val.String())
					}
				default:
					switch val.Interface().(type) {
					case time.Time:
						self.Params.Filter(typ.Tag.Get("field"), val.Interface())
					}

				}
			}
		}
	}

}

//查找数据
func (self *Object) All() ([]interface{}, error) {
	self.Lock()
	defer self.Unlock()
	self.autoWhere()
	if rows, err := self.Params.All(); err == nil {
		defer rows.Close()

		ret := []interface{}{}

		for rows.Next() {
			m := reflect.New(reflect.TypeOf(self.mode).Elem()).Elem()

			val := []interface{}{}
			for i := 0; i < m.NumField(); i++ {
				if name := m.Type().Field(i).Tag.Get("field"); len(name) > 0 {
					val = append(val, m.Field(i).Addr().Interface())
				}
			}
			rows.Scan(val...)
			ret = append(ret, m.Addr().Interface())
		}
		return ret, err
	} else {
		return nil, err
	}

}

//提取一个数据
func (self *Object) One() error {
	self.RLock()
	defer self.RUnlock()
	self.autoWhere()
	valMode := reflect.ValueOf(self.mode).Elem()
	typeMode := reflect.TypeOf(self.mode).Elem()
	vals := []interface{}{}
	for i := 0; i < valMode.NumField(); i++ {
		if name := typeMode.Field(i).Tag.Get("field"); len(name) > 0 {
			//vals[i] = valMode.Field(i).Addr().Interface()
			vals = append(vals, valMode.Field(i).Addr().Interface())
		}
	}
	err := self.Params.One(vals...)
	if err == nil {
		self.where = self.where[len(self.where):]
		return nil
	} else {
		return err
	}
}
