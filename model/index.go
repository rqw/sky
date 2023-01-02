package model

var GormModels = []interface{}{
	&User{},
	&Role{},
	&Menu{},
	&Api{},
	&OperationLog{}}

type ModelInterface interface {
	User | Role | Menu | Api | OperationLog
}
