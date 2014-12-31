// Copyright 2014 The 开心易点 All rights reserved.
// 非开源，公司内部程序， 不可转送带走，不可未经公司允许发布到公网、个人网盘、电子邮箱。不可做非公司电脑或非公司存储设备的备份。
// 不可对外泄露程序内部逻辑，程序结构，不可泄露程序数据结构以及相关加密算法。

// model解析支持

package orm

import (
	"errors"
	"reflect"
	"strings"
	"sync"
)

type Module interface {
	GetTableName() string
}

func Objects(mode Module) Object {
	param := new(Params)
	param.SetTable(mode.GetTableName())
	param.init()
	typ := reflect.TypeOf(mode).Elem()
	vals := []string{}
	for i := 0; i < typ.NumField(); i++ {
		if field := typ.Field(i).Tag.Get("field"); len(field) > 0 {
			vals = append(vals, field)
		}
	}
	param.SetField(vals...)
	obj := Object{}
	obj.mode = mode
	obj.params = param
	return obj
}

type Object struct {
	sync.RWMutex
	mode   Module
	params *Params
}

func (self *Object) Objects(mode Module) *Object {
	self.Lock()
	defer self.Unlock()
	param := new(Params)
	param.SetTable(mode.GetTableName())
	param.init()
	typ := reflect.TypeOf(mode).Elem()
	vals := []string{}
	for i := 0; i < typ.NumField(); i++ {
		if field := typ.Field(i).Tag.Get("field"); len(field) > 0 {
			vals = append(vals, field)
		}
	}
	param.SetField(vals...)
	self.mode = mode
	self.params = param
	return self
}

func (self *Object) Filter(name string, val interface{}) *Object {
	self.Lock()
	defer self.Unlock()
	typ := reflect.TypeOf(self.mode).Elem()
	fieldName := strings.Split(name, "__")
	if field, ok := typ.FieldByName(fieldName[0]); ok {
		name := field.Tag.Get("field")
		if len(fieldName) > 1 {
			name = name + "__" + fieldName[1]
		}
		self.params.Filter(name, val)
	}
	return self
}

func (self *Object) Filters(filters map[string]interface{}) *Object {
	self.Lock()
	defer self.Unlock()
	for k, v := range filters {
		self.Filter(k, v)
	}
	return self
}

func (self *Object) Limit(s, steq int) *Object {
	self.Lock()
	defer self.Unlock()
	self.params.limit = [2]int{s, steq}
	return self
}

func (self *Object) Db(name string) *Object {
	self.params.Db(name)
	return self
}

func (self *Object) Count() (int64, error) {
	self.RLock()
	defer self.RUnlock()
	return self.params.Count()
}
func (self *Object) Delete() (int64, error) {
	self.Lock()
	defer self.Unlock()
	valus := reflect.ValueOf(self.mode).Elem()
	for i := 0; i < valus.NumField(); i++ {
		typ := valus.Type().Field(i)
		val := valus.Field(i)
		if typ.Tag.Get("auto") == "true" {
			switch val.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if val.Int() > 0 {
					self.params.Filter(typ.Name, val.Int())
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if val.Uint() > 0 {
					self.params.Filter(typ.Name, val.Uint())
				}
			}
		}
	}
	if len(self.params.where) > 0 {
		res, err := self.params.Delete()
		if err != nil {
			return 0, err
		}
		return res.RowsAffected()
	} else {
		return 0, errors.New("No Where")
	}

}
func (self *Object) Save() (bool, int64, error) {
	self.Lock()
	defer self.Unlock()
	valus := reflect.ValueOf(self.mode).Elem()
	for i := 0; i < valus.NumField(); i++ {
		typ := valus.Type().Field(i)
		val := valus.Field(i)
		if typ.Tag.Get("auto") == "true" {
			switch val.Kind() {
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				if val.Int() > 0 {
					self.params.Filter(typ.Tag.Get("field"), val.Int())
					self.params.SetChange(typ.Tag.Get("field"), val.Interface())
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if val.Uint() > 0 {
					self.params.Filter(typ.Tag.Get("field"), val.Uint())
					self.params.SetChange(typ.Tag.Get("field"), val.Interface())
				}
			}
		} else {
			self.params.SetChange(typ.Tag.Get("field"), val.Interface())
		}
	}

	isNew, id, err := self.params.Save()
	return isNew, id, err
}

func (self *Object) All() ([]interface{}, error) {
	self.Lock()
	defer self.Unlock()
	if rows, err := self.params.All(); err == nil {
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

func (self *Object) One() error {
	self.RLock()
	defer self.RUnlock()
	if row := self.params.One(); row != nil {
		valMode := reflect.ValueOf(self.mode).Elem()
		typeMode := reflect.TypeOf(self.mode).Elem()
		vals := []interface{}{}
		for i := 0; i < valMode.NumField(); i++ {
			if name := typeMode.Field(i).Tag.Get("field"); len(name) > 0 {
				//vals[i] = valMode.Field(i).Addr().Interface()
				vals = append(vals, valMode.Field(i).Addr().Interface())
			}
		}
		err := row.Scan(vals...)
		return err
	}
	return errors.New("Get error")
}
