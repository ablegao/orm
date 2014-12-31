//Cache，用来沟通第一层cache 和同步数据到数据库

package orm

import (
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

func CacheMode(mode Module) (cache *CacheModule) {
	cache.mode = mode
	cache.Objects(mode)
	cache.cachekey = cache.GetCacheKey()

	return
}

type CacheModule struct {
	Cache
	Object
	cachekey string
}

func (self *CacheModule) Objects(mode Module) *CacheModule {

	self.Object.Objects(mode)
	self.cachekey = self.GetCacheKey()
	return self
}

func (self *CacheModule) GetCacheKey() string {

	value := reflect.ValueOf(self.mode).Elem()
	typeOf := reflect.TypeOf(self.mode).Elem()
	str := cache_prefix
	st := []byte("*")
	for i := 0; i < value.NumField(); i++ {
		field := typeOf.Field(i)
		if name := field.Tag.Get("cache"); len(name) > 0 {
			val := value.Field(i)
			if prefix := field.Tag.Get("cache_prefix"); len(prefix) > 0 {
				str = append(str, []byte(prefix+":")...)
			}
			str = append(str, []byte(":")...)
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
			str = append(str, []byte(":"+name)...)
		}
	}
	return string(str)
}

func (self CacheModule) Incrby(key string, val int64) (ret int64, err error) {
	ret, err = self.Cache.Hincrby(self.cachekey, key, val)
	if val > 0 {
		self.Object.Change(key+"_add", val)
	} else if val < 0 {
		self.Object.Change(key+"_sub", val)
	}
	go self.Object.Save()
	return
}

func (self CacheModule) Incry(key string) (val int64, err error) {
	val, err = self.Incrby(key, 1)

	return
}

func (self CacheModule) Set(key string, val interface{}) (err error) {
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
	if err != nil {
		return
	}

	go self.Object.Change(key, val).Save()

	return
}

func (self *CacheModule) Filter(name string, val interface{}) *CacheModule {

	return self
}

func (self *CacheModule) Filters(filters map[string]interface{}) *CacheModule {
	return self
}

func (self *CacheModule) Change(name string, val interface{}) *CacheModule {
	return self
}

func (self *CacheModule) Orderby(name ...string) *CacheModule {
	return self
}

func (self *CacheModule) Limit(page, step int) *CacheModule {
	return self
}
func (self *CacheModule) Count() (int64, error) {
	return 0, nil
}
func (self *CacheModule) Delete() (int64, error) {
	return 0, nil
}
func (self *CacheModule) Save() (isnew bool, id int64, err error) {
	go self.Object.Save()
	return
}
func (self *CacheModule) All() ([]interface{}, error) {
	return nil, nil
}

func (self *CacheModule) One() error {
	return nil
}
