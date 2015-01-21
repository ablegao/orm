##About 

一个数据库ORM.

## How to use?

### Insert 
go get github.com/ablegao/orm


##数据库Model 建立方法

    //引用模块
    import "github.com/ablegao/orm"

    //mysql 驱动
    import _ "github.com/go-sql-driver/mysql"
    
    //建立连接 
    // 参数分别为 名称 ， 驱动， 连接字符串
    // 注：必须包含一个default 连接， 作为默认连接。
    orm.NewDatabase("default" , "mysql" , "user:passwd@ip/database?charset=utf8")


    //建立一个数据模型。 
	type UserInfo struct {
		orm.Object
		Id int64 `field:"id" auto:"true" index:"pk"`
		Name string `field:"username"`
		Passwd string `field:"password"`
	}

[更多信息>>](docs/mode.md)

##新增 CacheModel 模型， 支持分布式redis作为数据库缓存。 

import "github.com/ablegao/orm"
import _ "github.com/go-sql-driver/mysql"

type userB struct {
	CacheModule
	Uid     int64  `field:"Id" index:"pk" cache:"user" `
	Alias   string `field:"Alias"`
	Money int64  `field:"money"	`
}

func main(){
	orm.CacheConsistent.Add("127.0.0.1:6379")  //添加多个redis服务器
	orm.SetCachePrefix("nado") //默认nado .  将作为redis key 的前缀
	NewDatabase("default", "mysql", "happy:passwd@tcp(127.0.0.1:3306)/mydatabase?charset=utf8")


	b := new(userB)
	b.Uid = 10000
	err:=b.Objects(b).One()
	if err!= nil {
		panic(err)
	}
	fmt.Println(b.Uid ,b.Alias ,b.Money)

	b.Incrby("Money" , 100)
	fmt.Println(b.Money)
	b.Save() //不执行不会保存到数据库 只会修改redis数据。 


}