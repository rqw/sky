package common

import (
	"sky/config"
	"sky/model"
	"sky/util"

	"reflect"

	"github.com/thoas/go-funk"
	"github.com/wxnacy/wgo/arrays"
)

func getItemIdVal(item interface{}) uint64 {
	values := reflect.ValueOf(item)
	fieldValue := values.FieldByName("ID")
	return fieldValue.Uint()
}
func initDataByType[T interface{}](conf string) ([]*T, []*T) {
	dbRecode := make([]*T, 0)
	fileRecode := make([]*T, 0)
	newRecode := make([]*T, 0)
	// 1.写入角色数据
	util.ConfFileToModel(conf, &fileRecode)
	ids := make([]uint64, len(fileRecode))
	for i, item := range fileRecode {
		ids[i] = getItemIdVal(item)
	}
	res := DB.Find(&dbRecode, ids)

	//数据库中已经存在角色记录不再写入
	if res.RowsAffected > 0 {
		ids = make([]uint64, res.RowsAffected)
		for i, item := range dbRecode {
			ids[i] = getItemIdVal(item)
		}
		for _, item := range fileRecode {
			if arrays.ContainsUint(ids, getItemIdVal(item)) == -1 {
				newRecode = append(newRecode, item)
			}
		}
	} else {
		newRecode = append(newRecode, fileRecode...)
	}
	if len(newRecode) > 0 {
		for _, item := range newRecode {
			err := DB.Create(&item).Error
			if err != nil {
				Log.Errorf("insert %s  data fail ：%v", conf, err)
			}
		}
	}

	return newRecode, fileRecode
}

// 初始化数据库数据
func InitData() {
	// 是否初始化数据
	if !config.Conf.System.InitData {
		return
	}
	_, roles := initDataByType[model.Role]("role")
	initDataByType[model.Menu]("menu")
	initDataByType[model.User]("user")
	newApi, _ := initDataByType[model.Api]("api")
	newRoleCasbin := make([]model.RoleCasbin, 0)
	for _, api := range newApi {

		// 管理员拥有所有API权限
		newRoleCasbin = append(newRoleCasbin, model.RoleCasbin{
			Keyword: roles[0].Keyword,
			Path:    api.Path,
			Method:  api.Method,
		})

		// 非管理员拥有基础权限
		basePaths := []string{
			"/base/login",
			"/base/logout",
			"/base/refreshToken",
			"/user/info",
			"/menu/access/tree/:userId",
		}

		if funk.ContainsString(basePaths, api.Path) {
			newRoleCasbin = append(newRoleCasbin, model.RoleCasbin{
				Keyword: roles[1].Keyword,
				Path:    api.Path,
				Method:  api.Method,
			})
			newRoleCasbin = append(newRoleCasbin, model.RoleCasbin{
				Keyword: roles[2].Keyword,
				Path:    api.Path,
				Method:  api.Method,
			})
		}
	}
	if len(newRoleCasbin) > 0 {
		rules := make([][]string, 0)
		for _, c := range newRoleCasbin {
			rules = append(rules, []string{
				c.Keyword, c.Path, c.Method,
			})
		}
		isAdd, err := CasbinEnforcer.AddPolicies(rules)
		if !isAdd {
			Log.Errorf("写入casbin数据失败：%v", err)
		}
	}
}
