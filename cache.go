// Copyright 2014 The 开心易点 All rights reserved.
// 非开源，公司内部程序， 不可转送带走，不可未经公司允许发布到公网、个人网盘、电子邮箱。不可做非公司电脑或非公司存储设备的备份。
// 不可对外泄露程序内部逻辑，程序结构，不可泄露程序数据结构以及相关加密算法。

//Cache，用来沟通第一层cache 和同步数据到数据库

package orm

import (
	"fmt"
	"reflect"
	"strconv"
)

var cache_prefix []byte = []byte("nado")

func SetCachePrefix(str string) {
	cache_prefix = []byte(str)
}

type Cache interface {
	Set(key string, b []byte) error
	Get(key string) ([]byte, error)
	Keys(key string) ([]string, error)
	//Incy(key string) (int64, error)
	Incrby(key string, n int64) (int64, error)
	Hset(key, field string, b []byte) (bool, error)
	Hmset(key string, maping interface{}) error
	Hget(key, field string) ([]byte, error)
	Hincrby(key, filed string, n int64) (int64, error)
	Exists(key string) (bool, error)
	Del(key string) (bool, error)
	GetConnect()
}

func CacheMode(mode Module) (cache *CacheObj) {
	obj := Objects(mode)
	obj.One()
	cache.mode = mode

	cache.Objects(mode)
	cache.cachekey = cache.GetCacheKey()

	if has, err := cache.Cache.Exists(cache.cachekey); err == nil {
		if has {

		} else {

		}
	}

	return
}

type CacheObj struct {
	Cache
	*Object
	cachekey string
}

func (self *CacheObj) Objects(mode Module) *CacheObj {
	self.Object = new(Object)
	self.Object.Objects(mode)

	return self
}

func (self *CacheObj) GetCacheKey() string {

	value := reflect.ValueOf(self.mode).Elem()
	typeOf := reflect.TypeOf(self.mode).Elem()
	str := cache_prefix
	st := []byte("*")
	for i := 0; i < value.NumField(); i++ {
		field := typeOf.Field(i)
		if name := field.Tag.Get("cache"); len(name) > 0 {
			val := value.Field(i)
			str = append(str, []byte(name+":")...)
			switch field.Type.Kind() {
			case reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uint8, reflect.Uint16:
				if val.Uint() <= 0 {
					str = append(str, st...)
				} else {
					str = strconv.AppendUint(str, val.Uint(), 10)
				}
			case reflect.Int32, reflect.Int64, reflect.Int, reflect.Int8, reflect.Int16:
				if val.Int() <= 0 {
					str = append(str, st...)
				} else {
					str = strconv.AppendInt(str, val.Int(), 10)
				}
			case reflect.String:
				if val.Len() <= 0 {
					str = append(str, st...)
				} else {
					str = append(str, []byte(val.String())...)
				}
			case reflect.Bool:
				switch val.Bool() {
				case true:
					str = append(str, []byte("true")...)
				case false:
					str = append(str, []byte("false")...)
				}
			}
		}
	}
	return string(str)
}

func (self *CacheObj) Incrby(key string, val int64) (ret int64, err error) {
	ret, err = self.Cache.Hincrby(self.cachekey, key, val)
	if val > 0 {
		str := fmt.Sprintf("+%d", val)
		self.Object.params.SetChange(key, str)
	} else if val < 0 {
		self.Object.params.SetChange(key, val)
	}
	return
}

func (self *CacheObj) Incry(key string) (val int64, err error) {
	val, err = self.Incrby(key, 1)
	return
}

func (self *CacheObj) Set(key string, val interface{}) (err error) {
	b := []byte{}
	switch val.(type) {
	case uint32, uint64, uint16, uint8:
		b = strconv.AppendUint(b, reflect.ValueOf(val).Uint(), 10)
	case string:
		b = append(b, []byte(val.(string))...)
	case int32, int64, int16, int8:
		b = strconv.AppendInt(b, reflect.ValueOf(val).Int(), 10)
	case bool:
		b = strconv.AppendBool(b, val.(bool))
	}
	_, err = self.Cache.Hset(self.cachekey, key, b)

	self.Object.params.SetChange(key, val)
	return
}

func (self *CacheObj) Save() (err error) {
	go self.Object.Save()
	return
}
