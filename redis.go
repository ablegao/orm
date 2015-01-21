package orm

import (
	"time"

	"github.com/hoisie/redis"
)

var CacheConsistent = NewConsistent()

var (
	getRedis = make(chan *comandGetRedisConn)

	updateRedis = make(chan []string)
	RedisServer = map[string]*redis.Client{}
)

//获取目标地址的功能
type comandGetRedisConn struct {
	Key  string
	Call chan *redis.Client
}

func init() {
	go goRedisRuntime()

}

//守护服务
func goRedisRuntime() {
	for {
		select {
		case mapping := <-updateRedis:
			CacheConsistent.Set(mapping)
		case t := <-getRedis:
			addr, err := getRedisAddrByKey(t.Key)
			if err != nil {
				t.Call <- nil
				return
			}

			client, ok := RedisServer[addr]
			if !ok {
				client = new(redis.Client)
				client.Addr = addr
				client.MaxPoolSize = 8
				RedisServer[addr] = client
			}
			t.Call <- client

		}
	}
}

//通过一致性hash服务， 得到当前key应该分配给哪个redis服务器
func getRedisAddrByKey(key string) (string, error) {
	return CacheConsistent.Get(key)
}

func GetRedisClient(key string) *redis.Client {
	p := new(comandGetRedisConn)
	p.Call = make(chan *redis.Client, 1)
	p.Key = key
	getRedis <- p
	select {
	case item := <-p.Call:
		return item
	case <-time.After(time.Second * 5):
		return nil
	}
}
