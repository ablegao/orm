//Cache，用来沟通第一层cache 和同步数据到数据库

package orm

import (
	"errors"
	"log"
	"os"
	"reflect"
	"sort"
	"strconv"
	"time"
)

var Debug = log.New(os.Stdout, "ORM-DEBUG", log.Lshortfile|log.LstdFlags)
var Error = log.New(os.Stdout, "ORM-ERROR", log.Lshortfile|log.LstdFlags)
var cache_prefix []byte = []byte("nado")
var cache_db = 0
var st []byte = []byte("*")

var (
	ErrKeyNotExist = errors.New("keys not exists")
)

func SetCachePrefix(str string) {
	cache_prefix = []byte(str)

}
func SetDefaultCacheDb(db int) {
	cache_db = db
}

var debug_sql bool = false

func SetDebug(b bool) {
	debug_sql = b
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
}

func CacheMode(mode Module) (cache *CacheModule) {
	cache = new(CacheModule)
	cache.CacheFileds = []string{}
	cache.CacheNames = []string{}
	cache.Object.Objects(mode)
	typeOf := reflect.TypeOf(cache.mode).Elem()
	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		if name := field.Tag.Get("cache"); len(name) > 0 {
			cache.CacheFileds = append(cache.CacheFileds, field.Tag.Get("field"))
			cache.CacheNames = append(cache.CacheNames, name)
		}
		if prefix := field.Tag.Get("cache_prefix"); len(prefix) > 0 {
			cache.cache_prefix = prefix
		}
	}
	cache.Cache = GetRedisClient("default")
	return
}

type CacheModuleInteerface interface {
	Objects(Module) CacheModuleInteerface
	Ca(interface{}) CacheModuleInteerface //一致性hash 默认处理方式
	Db(string) CacheModuleInteerface      //数据库连接
	Filter(name string, val interface{}) CacheModuleInteerface
	GetCacheKey() string
	Incrby(string, int64) (int64, error)
	Incry(string) (int64, error)
	Set(string, interface{}) error
	Save() (bool, int64, error)
	One() error
	SaveToCache() error
	All() ([]interface{}, error)
	AllCache() ([]interface{}, error)
}

type CacheModule struct {
	Cache
	Object
	cachekey      string
	CacheFileds   []string
	CacheNames    []string
	cache_prefix  string
	cache_address string
}

func (self *CacheModule) Objects(mode Module) *CacheModule {
	self.CacheFileds = []string{}
	self.CacheNames = []string{}
	self.Object.Objects(mode)
	typeOf := reflect.TypeOf(self.mode).Elem()
	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		if name := field.Tag.Get("cache"); len(name) > 0 {
			self.CacheFileds = append(self.CacheFileds, field.Tag.Get("field"))
			self.CacheNames = append(self.CacheNames, name)
		}
		if prefix := field.Tag.Get("cache_prefix"); len(prefix) > 0 {
			self.cache_prefix = prefix
		}

	}
	self.Cache = GetRedisClient("default")
	return self
}

func (self *CacheModule) Db(name string) *CacheModule {
	self.Params.Db(name)
	return self
}

func GetCacheConn(key interface{}) (address string, c Cache) {
	value := reflect.ValueOf(key)
	typeOf := reflect.TypeOf(key)
	b := []byte{}
	switch typeOf.Kind() {
	case reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uint8, reflect.Uint16:
		b = strconv.AppendUint(b, value.Uint(), 10)
	case reflect.Int32, reflect.Int64, reflect.Int, reflect.Int8, reflect.Int16:
		b = strconv.AppendInt(b, value.Int(), 10)
	case reflect.Float32, reflect.Float64:
		b = strconv.AppendFloat(b, value.Float(), 'f', 0, 64)
	case reflect.String:
		b = append(b, []byte(value.String())...)
	case reflect.Bool:
		b = strconv.AppendBool(b, value.Bool())
	}
	address = string(b)
	c = GetRedisClient(address)
	return
}

func (self *CacheModule) Ca(key interface{}) *CacheModule {
	self.cache_address, self.Cache = GetCacheConn(key)
	return self
}

func (self *CacheModule) GetCacheKey() string {

	value := reflect.ValueOf(self.mode).Elem()
	typeOf := reflect.TypeOf(self.mode).Elem()
	str := cache_prefix
	str = append(str, []byte(self.cache_prefix)...)

	for i := 0; i < value.NumField(); i++ {
		field := typeOf.Field(i)
		self.CacheFileds = append(self.CacheFileds, field.Name)
		if name := field.Tag.Get("cache"); len(name) > 0 {
			val := value.Field(i)
			str = append(str, []byte(":")...)
			str = append(str, self.fieldToByte(val.Interface())...)
			str = append(str, []byte(":"+name)...)
		}
	}
	return string(str)
}
func (self *CacheModule) setModeFieldUint(field string, val interface{}) {

	value := reflect.ValueOf(self.mode).Elem()
	item := value.FieldByName(field)
	switch item.Type().Kind() {
	case reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uint8, reflect.Uint16:
		item.SetUint(uint64(val.(int64)))
	case reflect.Int32, reflect.Int64, reflect.Int, reflect.Int8, reflect.Int16:
		item.SetInt(val.(int64))
	default:
		Error.Println("(CacheModule) setModeFieldUint", field, item.Type().Kind())
	}
}

func (self *CacheModule) Incrby(key string, val int64) (ret int64, err error) {
	if self.cachekey == "" {
		self.cachekey = self.GetCacheKey()
	}
	ret, err = self.Cache.Hincrby(self.cachekey, key, val)

	if val >= 0 {
		self.Object.Change(key+"__add", val)
	} else if val < 0 {
		self.Object.Change(key+"__sub", val)
	}
	self.setModeFieldUint(key, ret)
	//go self.Object.Save()
	return
}

func (self *CacheModule) Incry(key string) (val int64, err error) {
	val, err = self.Incrby(key, 1)

	return
}

func (self *CacheModule) Set(key string, val interface{}) (err error) {

	b := []byte{}
	field := reflect.ValueOf(self.mode).Elem().FieldByName(key)
	switch val.(type) {
	case uint32, uint64, uint16, uint8, uint:
		val := reflect.ValueOf(val).Uint()
		b = strconv.AppendUint(b, val, 10)
		field.SetUint(val)
	case string:
		b = append(b, []byte(val.(string))...)
		field.SetString(val.(string))

	case int32, int64, int16, int8, int:
		val := reflect.ValueOf(val).Int()
		b = strconv.AppendInt(b, val, 10)
		field.SetInt(val)
	case float32, float64:
		val := reflect.ValueOf(val).Float()
		b = strconv.AppendFloat(b, val, 'f', 0, 64)
		field.SetFloat(val)
	case bool:
		b = strconv.AppendBool(b, val.(bool))
		field.SetBool(val.(bool))
	case time.Time:
		field.Set(reflect.ValueOf(val))
		b = append(b, []byte(val.(time.Time).Format(time.RFC1123Z))...)
	default:
		Error.Println("undefined val type")
	}
	if self.cachekey == "" {
		self.cachekey = self.GetCacheKey()
	}

	var over bool
	over, err = self.Cache.Hset(self.cachekey, key, b)
	if err != nil {
		return
	}
	if over == false {
		return errors.New(self.cachekey + " hset " + key + " error !")
	}
	self.Object.Change(key, val)
	//go self.Object.Change(key, val).Save()

	return
}

func (self *CacheModule) Save() (isnew bool, id int64, err error) {

	key := self.GetCacheKey()
	if i, err := self.Exists(key); err == nil && i == false {
		self.SaveToCache()
	}
	return self.Object.Save()
}

func (self *CacheModule) Filter(name string, val interface{}) *CacheModule {
	self.Object.Filter(name, val)
	return self
}
func (self *CacheModule) Orderby(order ...string) *CacheModule {
	self.Object.Orderby(order...)
	return self
}
func (self *CacheModule) Limit(page, step int) *CacheModule {
	self.Object.Limit(page, step)
	return self
}
func (self *CacheModule) AllOnCache() ([]interface{}, error) {
	key := self.getKey()
	if keys, err := self.Keys(key); err == nil && len(keys) > 0 {
		//(keys)
		sort.Sort(sort.StringSlice(keys))
		if self.limit != NULL_LIMIT {
			page := (self.limit[0] - 1) * self.limit[1]
			step := self.limit[0] * self.limit[1]
			if step > len(keys) {
				step = len(keys)
			}
			if page < len(keys) {
				keys = keys[page:step]
				vals := make([]interface{}, len(keys))
				for i, k := range keys {
					vals[i] = self.key2Mode(k)
				}
				return vals, nil
			} else {
				return make([]interface{}, 0), nil
			}
		} else {
			vals := make([]interface{}, len(keys))
			for i, k := range keys {
				vals[i] = self.key2Mode(k)
			}
			return vals, nil
		}
	} else {
		if err == nil {
			return []interface{}{}, nil
		}
		return nil, err
	}
}
func (self *CacheModule) All() ([]interface{}, error) {
	if rets, err := self.AllOnCache(); err == nil && len(rets) > 0 {
		return rets, err
	} else {
		if debug_sql {
			Error.Println("all error:", err, len(rets))
		}
		//self.Object.All()
		if rets, err := self.Object.All(); err == nil {
			for _, item := range rets {

				//Debug.Println(item)
				self.saveToCache(item.(Module))
			}
			return rets, nil
		} else {
			return nil, err
		}

	}
}

func (self *CacheModule) DeleteOnCache() error {
	if len(self.cachekey) <= 0 {
		self.cachekey = self.getKey()
	}
	if _, err := self.Del(self.cachekey); err != nil {
		return err
	}

	return nil
}
func (self *CacheModule) OneOnCache() error {
	key := self.getKey()
	self.cachekey = key

	n, err := self.Exists(key)
	if err != nil {
		if debug_sql {
			Debug.Println(err)
		}
		return err
	}
	if debug_sql {
		Debug.Println("keys ", self.cachekey, " is exists ", n)
	}
	if n == false {
		return ErrKeyNotExist
	}

	self.where = self.where[len(self.where):]
	val := reflect.ValueOf(self.mode).Elem()
	typ := reflect.TypeOf(self.mode).Elem()
	for i := 0; i < val.NumField(); i++ {
		if b, err := self.Cache.Hget(key, typ.Field(i).Name); err == nil {
			switch val.Field(i).Kind() {
			case reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uint8, reflect.Uint16:
				id, _ := strconv.ParseUint(string(b), 10, 64)
				val.Field(i).SetUint(id)
			case reflect.Int32, reflect.Int64, reflect.Int, reflect.Int8, reflect.Int16:
				id, _ := strconv.ParseInt(string(b), 10, 64)
				val.Field(i).SetInt(id)
			case reflect.Float32, reflect.Float64:
				id, _ := strconv.ParseFloat(string(b), 64)
				val.Field(i).SetFloat(id)
			case reflect.String:
				val.Field(i).SetString(string(b))
			case reflect.Bool:
				id, _ := strconv.ParseBool(string(b))
				val.Field(i).SetBool(id)
			default:
				switch val.Field(i).Interface().(type) {
				case time.Time:
					if time, e := time.Parse(time.RFC1123Z, string(b)); e == nil {
						val.Field(i).Set(reflect.ValueOf(time))
					}
				}
			}
			if index := typ.Field(i).Tag.Get("index"); len(index) > 0 {
				self.Object.Filter(typ.Field(i).Name, val.Field(i).Interface())
			}
		}
	}
	self.hasRow = true
	return nil
}

func (self *CacheModule) One() error {

	if err := self.OneOnCache(); err != nil {
		//return errors.New("keys " + key + " not exists!")
		if debug_sql {
			Debug.Println("keys ", self.cachekey, "error", err)
		}
		err = self.Object.One()
		if err == nil {
			defer self.SaveToCache()
		} else {
			Error.Println(err)
		}
		return err
	}
	return nil
}
func (self *CacheModule) CountOnCache() (int64, error) {
	if keys, err := self.Cache.Keys(self.getKey()); err == nil {
		return int64(len(keys)), nil
	} else {
		return 0, err
	}
}
func (self *CacheModule) Count() (int64, error) {
	if keys, err := self.Cache.Keys(self.getKey()); err == nil && len(keys) > 0 {
		return int64(len(keys)), nil
	}
	return self.Object.Count()
}

func (self *CacheModule) getKey() string {
	key := ""
	if len(self.where) > 0 {
		key = self.where2Key()
	} else {
		key = self.GetCacheKey()
	}
	return key
}
func (self *CacheModule) where2Key() string {
	str := cache_prefix
	str = append(str, []byte(self.cache_prefix)...)
	for index, field := range self.CacheFileds {
		for _, wh := range self.where {
			if wh.name == field {
				str = append(str, []byte(":")...)
				str = append(str, self.fieldToByte(wh.val)...)
				str = append(str, []byte(":"+self.CacheNames[index])...)
				goto NEXT
			}

		}
		str = append(str, []byte(":*:"+self.CacheNames[index])...)
	NEXT:
	}
	return string(str)
}

func (self *CacheModule) fieldToByte(value interface{}) (str []byte) {
	typ := reflect.TypeOf(value)
	val := reflect.ValueOf(value)

	switch typ.Kind() {
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
	case reflect.Float32, reflect.Float64:
		if val.Float() == 0.0 {
			str = append(str, st...)
		} else {
			str = strconv.AppendFloat(str, val.Float(), 'f', 0, 64)
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
	default:
		switch value.(type) {
		case time.Time:
			str = append(str, []byte(value.(time.Time).Format(time.RFC1123Z))...)
		}
	}
	return
}

func (self *CacheModule) key2Mode(key string) interface{} {
	typ := reflect.TypeOf(self.mode).Elem()
	val := reflect.New(typ).Elem()
	for i := 0; i < typ.NumField(); i++ {
		if b, err := self.Cache.Hget(key, typ.Field(i).Name); err == nil {
			switch val.Field(i).Kind() {
			case reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uint8, reflect.Uint16:
				id, _ := strconv.ParseUint(string(b), 10, 64)
				val.Field(i).SetUint(id)
			case reflect.Int32, reflect.Int64, reflect.Int, reflect.Int8, reflect.Int16:
				id, _ := strconv.ParseInt(string(b), 10, 64)
				val.Field(i).SetInt(id)
			case reflect.Float32, reflect.Float64:
				id, _ := strconv.ParseFloat(string(b), 64)
				val.Field(i).SetFloat(id)
			case reflect.String:
				val.Field(i).SetString(string(b))
			case reflect.Bool:
				id, _ := strconv.ParseBool(string(b))
				val.Field(i).SetBool(id)
			default:
				switch val.Field(i).Interface().(type) {
				case time.Time:
					//str = append(str, []byte(values.(time.Time).Format(time.RFC1123Z))...)
					if time, e := time.Parse(time.RFC1123Z, string(b)); e == nil {
						val.Field(i).Set(reflect.ValueOf(time))
					}
				}
			}
		}
	}
	return val.Addr().Interface()
}
func (self CacheModule) saveToCache(mode Module) error {
	return CacheMode(mode).Ca(self.cache_address).SaveToCache()
}

func (self *CacheModule) SaveToCache() error {
	key := self.GetCacheKey()
	maping := map[string]interface{}{}
	vals := reflect.ValueOf(self.mode).Elem()
	typ := reflect.TypeOf(self.mode).Elem()
	for i := 0; i < vals.NumField(); i++ {
		field := typ.Field(i)
		if name := field.Tag.Get("field"); len(name) > 0 {
			if nocache := field.Tag.Get("no_cache"); len(nocache) == 0 {
				switch vals.Field(i).Interface().(type) {
				case time.Time:
					maping[field.Name] = vals.Field(i).Interface().(time.Time).Format(time.RFC1123Z)
				default:
					maping[field.Name] = vals.Field(i).Interface()
				}
			}
		}

		//补充一个仅存在于cache中的字段。
		if name := field.Tag.Get("cache_only_field"); len(name) > 0 {
			switch vals.Field(i).Interface().(type) {
			case time.Time:
				maping[field.Name] = vals.Field(i).Interface().(time.Time).Format(time.RFC1123Z)
			default:
				maping[field.Name] = vals.Field(i).Interface()
			}
		}
	}
	return self.Hmset(key, maping)
}
