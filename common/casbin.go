package common

import (
	"sky/config"

	"github.com/casbin/casbin/v2"
	gormadapter "github.com/casbin/gorm-adapter/v3"
)

// 全局CasbinEnforcer
var CasbinEnforcer *casbin.Enforcer

// 初始化casbin策略管理器
func InitCasbinEnforcer() bool {
	a, err := gormadapter.NewAdapterByDB(DB)
	if LogErr("InitCasbinEnforcer fail in NewAdapterByDB：%v", err) {
		return false
	}
	e, err := casbin.NewEnforcer(config.Conf.Casbin.ModelPath, a)
	if LogErr("InitCasbinEnforcer fail in NewEnforcer：%v", err) {
		return false
	}

	err = e.LoadPolicy()
	if LogErr("InitCasbinEnforcer fail in LoadPolicy%v", err) {
		return false
	}
	CasbinEnforcer = e
	Log.Info("InitCasbinEnforcer success!")
	return true

}
