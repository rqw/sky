package common

import (
	"sky/config"
	"sky/model"
	"sky/util"

	"reflect"

	"github.com/thoas/go-funk"
)

// 通过反射获取ID字段值
func getItemIdVal(item interface{}) uint64 {
	values := reflect.ValueOf(item)
	fieldValue := values.FieldByName("ID")
	return fieldValue.Uint()
}

// 根据传入类型初始化数据
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
			if funk.ContainsUInt64(ids, getItemIdVal(item)) {
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
	_, menus := initDataByType[model.Menu]("menu")
	initDataByType[model.User]("user")
	newApi, _ := initDataByType[model.Api]("api")
	rules := make([][]string, 0)
	casbinCache := make(map[string]map[uint]model.Api)
	roleMap := make(map[uint]string)
	apiMap := make(map[uint]model.Api)
	for _, role := range roles {
		roleMap[role.ID] = role.Keyword
	}
	for _, api := range newApi {
		apiMap[api.ID] = *api
	}

	for _, menu := range menus {
		for _, role := range menu.Roles {
			roleCache := casbinCache[roleMap[role.ID]]
			if roleCache != nil {
				casbinCache[roleMap[role.ID]] = make(map[uint]model.Api)
				roleCache = casbinCache[roleMap[role.ID]]
			}
			for _, api := range menu.Apis {
				roleCache[api.ID] = apiMap[api.ID]
			}
		}
	}
	for keyword, item := range casbinCache {
		for _, api := range item {
			rules = append(rules, []string{
				keyword, api.Path, api.Method,
			})
		}
	}
	if len(rules) > 0 {
		isAdd, err := CasbinEnforcer.AddPolicies(rules)
		if !isAdd {
			Log.Errorf("写入casbin数据失败：%v", err)
		}
	}
}
