
###  例子
    
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
		Id int64 `field:"id" index:"pk"`
		Name string `field:"username"`
		Passwd string `field:"password"`
	}

    //数据库表名称
	func(self *UserInfo) GetTableName()string{

		return "database.user_info"
	}

	//查询一个用户名为 "test1"的账户  
	user:=new(UserInfo)
	err:=user.Objects(user).Filter("Name","test1").One()
	fmt.Println(user.Id , user.Passwd , user.Name)

	//Update 
	user.Name="test2"
	user.Objects(user).Save()
	// or 
	user.Objects(user).Filter("Id" , 1).Change("Name" , "test2").Save()


    //查询id小于10的所有数据

    users , err:=user.Objects(user).Filter("Id__lt",10).All()
    if err == nil {
        for _,userinfo:= range users{
        	u:=userinfo.(*UserInfo)
            fmt.Println(u.Id , u.Passwd , u.Name)
        }
    }

    //Create 
    user:=new(UserInfo)
    user.Name ="test1"
    user.Passwd ="123456"
    id  , err:=user.Objects(user).Save()


    
    //delete
    user.Objects(user).Delete()
    
    // User other Database connect 
    orm.NewDatabase("other" , "mysql" , "user:passwd@ip/database?charset=utf8")
    user.Objects(user).Db("other").Filter(x ,x).Delete()
    // or 
    user.Objects(user).Filter().Db("other").XXX()

## Filter or FilterOr
.Filter(fieldname , val )

Filter 作为orm 的主要作用是过滤查询条件， 最终将会转换为sql 语句中的where 条件语句。 可以填写多次， 多次数据为and 关系

FilterOr 作为Orm 的主要过滤查询条件， 最终将妆化为sql 语句的where 条件语句 , 可以填写多次， 多次数据以 or 连接

user.Objects(user).Filter("Name" , "test1").FilterOr("Name" , "test2").All()
//select id,username,passwd from database.user_info where username='test1' or username='test2'

##关于Filter字段的魔法参数

###目前支持:	
	__exact        精确等于 like 'aaa'
	 __iexact    精确等于 忽略大小写 ilike 'aaa'
	 __contains    包含 like '%aaa%'
	 __icontains    包含 忽略大小写 ilike '%aaa%'，但是对于sqlite来说，contains的作用效果等同于icontains。
	__gt    大于
	__gte    大于等于
	__ne    不等于
	__lt    小于
	__lte    小于等于
	__startswith   以...开头
	__istartswith   以...开头 忽略大小写
	__endswith     以...结尾
	__iendswith    以...结尾，忽略大小写

###尚未支持:
	__in     存在于一个list范围内
	__range    在...范围内
	__year       日期字段的年份
	__month    日期字段的月份
	__day        日期字段的日
	__isnull=True/False


##Change

修改数据， 执行时，相当于 sql 语句中的set 

传入一个结构字段 和值
.Change("Field" , 1)

update from xxx set field=1 

可以添加魔法参数：

.Change("Field__add" ,1 )
update from xxx set field=field+1

__add 累加 field=field+1
__sub 累减 field=field-1
__mult 累乘 field=field*1
__div 累计出发 field=field/1
