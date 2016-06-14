package main

import (
	"fmt"
	"github.com/tangyang/simple-http-server/config"
	"github.com/tangyang/simple-http-server/dao"
	tpprof "github.com/tangyang/simple-http-server/pprof"
	"os"
	"os/signal"
	"runtime/pprof"
	"syscall"
)

func main() {

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	tpprof.InitPprof()

	// conf := new(config.Config)
	conf := config.NewConfig()

	if conf.InitDB {
		userDao := &dao.UserDao{}
		relationDao := &dao.RelationDao{}
		err := userDao.CreateUserSchema(conf)
		if err != nil {
			fmt.Printf("Fail to create user schema, error: %s\n", err.Error())
			return
		}
		err = relationDao.CreateRelationSchema(conf)
		if err != nil {
			fmt.Printf("Fail to create relation schema, error: %s\n", err.Error())
		}
		fmt.Println("Init database schema... ")
		return
	}

	initHttpServer(conf)

	for {
		s := <-signalChan

		fmt.Println("Program exiting, get a signal: ", s)
		if s == syscall.SIGQUIT {
			p := pprof.Lookup("heap")
			p.WriteTo(os.Stdout, 2)
		} else {
			os.Exit(0)
		}
	}
}
