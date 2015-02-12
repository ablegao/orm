package orm

type driversqlType func(param ParamsInterface) ModuleToSql

var driversql = map[string]driversqlType{
	"mysql":  func(param ParamsInterface) ModuleToSql { return MysqlModeToSql{param} },
	"sqlite": func(param ParamsInterface) ModuleToSql { return SqliteModeToSql{param} },
}

func NewMarsharlDriverSql(driverName string, fun driversqlType) {
	driversql[driverName] = fun
}

type ModuleToSql interface {
	Select() (sql string, val []interface{})
	Delete() (sql string, val []interface{})
	Update() (sql string, val []interface{})
	Insert() (sql string, val []interface{})
	Count() (sql string, val []interface{})
	Instance(ParamsInterface)
}
