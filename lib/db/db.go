package db

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/go-xorm/xorm"
	"github.com/kataras/golog"
	"github.com/kataras/iris"
	"github.com/pelletier/go-toml"
	"sync"
	"time"
)

var (
	engineMysqlGroup *xorm.EngineGroup
	engineMaster     *xorm.Engine
	engineSlave      *xorm.Engine
)

type DBEngine struct {
	Conf *toml.Tree
	Env  string //开发环境
}

func New(conf *toml.Tree, env string) *DBEngine {
	return &DBEngine{Conf: conf, Env: env}
}

//master
func (e *DBEngine) GetMaster() *xorm.Engine {
	return master(e.Conf)
}

func (e *DBEngine) GetGroup() *xorm.EngineGroup {
	if engineMysqlGroup != nil {
		return engineMysqlGroup
	}

	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()

	if engineMysqlGroup != nil {
		return engineMysqlGroup
	}

	masterEngine := master(e.Conf)
	slaveEngine := slave(e.Conf)
	engine, err := xorm.NewEngineGroup(masterEngine, []*xorm.Engine{slaveEngine})
	if err != nil {
		golog.Fatal("dbsource.engineGroup", err)
	}

	iris.RegisterOnInterrupt(func() {
		engine.Close()
	})

	err = engine.Ping()
	if err != nil {
		golog.Fatal("got err when ping db: ", err)
	}

	if e.Env == "prod" {
		engine.ShowSQL(false)
	} else {
		engine.ShowSQL(true)
	}
	timeLocation := e.Conf.Get("system.timeLocation").(string)
	var SysTimeLocation, _ = time.LoadLocation(timeLocation)
	engine.SetTZLocation(SysTimeLocation)
	// 性能优化的时候才考虑，加上本机的SQL缓存
	cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
	engine.SetDefaultCacher(cacher)
	engineMysqlGroup = engine
	golog.Info("dbgroup created ....")
	return engineMysqlGroup
}

func master(c *toml.Tree) *xorm.Engine {
	//golog.Debug("master: ",engineMaster)
	if engineMaster != nil {
		return engineMaster
	}

	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()
	//golog.Debug("master: ",engineMaster)
	if engineMaster != nil {
		return engineMaster
	}

	dbDriver := c.Get("db.drive").(string)
	dbHost := c.Get("db.master.host").(string)
	dbPort := c.Get("db.master.port").(string)
	dbUser := c.Get("db.master.user").(string)
	dbPwd := c.Get("db.master.pwd").(string)
	dbDbname := c.Get("db.master.dbname").(string)
	dbMaxIdleConns := int(c.Get("db.master.maxIdleConns").(int64))
	dbMaxOpenConns := int(c.Get("db.slave.maxOpenConns").(int64))

	driveSource := dbUser + ":" + dbPwd + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbDbname + "?charset=utf8"
	//fmt.Println(driveSource)
	engine, err := xorm.NewEngine(dbDriver, driveSource)
	if err != nil {
		golog.Fatal("dbsource.InstanceMaster", err)
	}
	//设置连接池的空闲数大小
	engine.SetMaxIdleConns(dbMaxIdleConns)
	//设置最大打开连接数
	engine.SetMaxOpenConns(dbMaxOpenConns)
	engineMaster = engine
	golog.Info("master db created ....")
	return engineMaster
}

func slave(c *toml.Tree) *xorm.Engine {

	if engineSlave != nil {
		return engineSlave
	}

	var lock sync.Mutex
	lock.Lock()
	defer lock.Unlock()

	if engineSlave != nil {
		return engineSlave
	}

	dbDriver := c.Get("db.drive").(string)
	dbHost := c.Get("db.slave.host").(string)
	dbPort := c.Get("db.slave.port").(string)
	dbUser := c.Get("db.slave.user").(string)
	dbPwd := c.Get("db.slave.pwd").(string)
	dbDbname := c.Get("db.slave.dbname").(string)
	dbMaxIdleConns := int(c.Get("db.slave.maxIdleConns").(int64))
	dbMaxOpenConns := int(c.Get("db.slave.maxOpenConns").(int64))

	driveSource := dbUser + ":" + dbPwd + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbDbname + "?charset=utf8"
	engine, err := xorm.NewEngine(dbDriver, driveSource)
	if err != nil {
		golog.Fatal("dbsource.InstanceSlave", err)
	}
	//设置连接池的空闲数大小
	engine.SetMaxIdleConns(dbMaxIdleConns)
	//设置最大打开连接数
	engine.SetMaxOpenConns(dbMaxOpenConns)
	golog.Info("slave db created ....")
	engineSlave = engine
	return engineSlave
}
