package service

import (
	"github.com/tangyang/simple-http-server/config"
	"github.com/tangyang/simple-http-server/dao"
	"github.com/tangyang/simple-http-server/model"
)

var userDao *dao.UserDao = &dao.UserDao{}

type UserService struct {
}

func (*UserService) AddUser(conf *config.Config, user *model.User) (bool, error) {
	b, err := userDao.AddUser(conf, user)
	return b, err
}

func (u *UserService) GetUserByName(conf *config.Config, name string) *model.User {
	return userDao.GetUserByName(conf, name)
}

func (u *UserService) GetAllUsers(conf *config.Config) []model.User {
	return userDao.GetAllUsers(conf)
}
