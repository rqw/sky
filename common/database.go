package common

import (
	"fmt"
	"sky/config"
	"sky/model"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// 全局mysql数据库变量
var DB *gorm.DB

// 初始化数据库
func InitDatabase() bool {
	var dialect gorm.Dialector
	switch config.Conf.System.DatabaseType {
	case "mysql":
		dialect = initMysqlDialect()
	case "sqlite":
		dialect = initSqliteDialect()
	default:
		Log.Panic("system.database-type: '%s' is not support.", config.Conf.System.DatabaseType)
		return false
	}
	//数据库描述信息拼装完成后执行
	if dialect != nil {
		DB = initDb(dialect)
		if DB != nil {
			dbAutoMigrate(DB)
		}
	}
	return DB != nil
}

// 初始化sqlite描述信息
func initSqliteDialect() gorm.Dialector {
	return sqlite.Open("sky.db")
}

// 初始化mysql描述信息
func initMysqlDialect() gorm.Dialector {
	dbConf := config.Conf.Database
	return mysql.Open(fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&%s",
		dbConf.Username,
		dbConf.Password,
		dbConf.Host,
		dbConf.Port,
		dbConf.Database,
		dbConf.Charset,
		dbConf.Collation,
		dbConf.Query,
	))
}

// 初始化数据库连接
func initDb(dialect gorm.Dialector) *gorm.DB {
	dbConf := config.Conf.Database
	db, err := gorm.Open(dialect, &gorm.Config{
		// 禁用外键(指定外键时不会在创建真实的外键约束)
		DisableForeignKeyConstraintWhenMigrating: true,
		//// 指定表前缀
		NamingStrategy: schema.NamingStrategy{TablePrefix: dbConf.TablePrefix},
	})
	if err != nil {
		Log.Panicf("initDb err: %v", err)
		panic(fmt.Errorf("initDb err: %v", err))
	}
	// 开启Debug日志
	if config.Conf.Database.LogMode {
		db.Debug()
	}
	sqlDB, err := db.DB()
	if err != nil {
		Log.Panicf("initDb err: %v", err)
		panic(fmt.Errorf("initDb err: %v", err))
	}
	if sqlDB != nil {
		// SetMaxIdleConns 设置空闲连接池中连接的最大数量
		sqlDB.SetMaxIdleConns(dbConf.MaxIdle)
		// SetMaxOpenConns 设置打开数据库连接的最大数量。
		sqlDB.SetMaxOpenConns(dbConf.MaxOpen)
		// SetConnMaxLifetime 设置了连接可复用的最大时间。
		sqlDB.SetConnMaxLifetime(time.Duration(dbConf.ConnMaxLifetime) * time.Second)
	}

	return db
}

// 自动迁移表结构
func dbAutoMigrate(DB *gorm.DB) {
	DB.AutoMigrate(model.GormModels...)
}
