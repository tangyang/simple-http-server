package dao

import (
	// _ "github.com/go-pg/pg"
	"github.com/tangyang/simple-http-server/config"
	pg "gopkg.in/pg.v4"
	"sync"
	"time"
)

var connector *PostgreConnector
var lock *sync.Mutex = &sync.Mutex{}

type PostgreConnector struct {
	Address  string
	DbName   string
	User     string
	Password string
	DB       *pg.DB
}

func NewPostgreConnector(conf *config.Config) *PostgreConnector {
	lock.Lock()
	defer lock.Unlock()
	if connector == nil {
		connector = &PostgreConnector{Address: conf.PgAddress, DbName: conf.PgDatabaseName, User: conf.PgUsername, Password: conf.PgPassword}
		connector.Connect(conf)
	}
	return connector
}

func (c *PostgreConnector) Connect(conf *config.Config) {
	readTimeout := time.Duration(conf.PgReadTimeout) * time.Second
	writeTimeout := time.Duration(conf.PgWriteTimeout) * time.Second
	idleTImeout := time.Duration(conf.PgIdleTimeout) * time.Second
	c.DB = pg.Connect(&pg.Options{Addr: c.Address, User: c.User, Password: c.Password,
		Database: c.DbName, ReadTimeout: readTimeout, WriteTimeout: writeTimeout,
		PoolSize: conf.PgPoolsize, IdleTimeout: idleTImeout})
}
