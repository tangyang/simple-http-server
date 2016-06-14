package dao

import (
	"github.com/tangyang/simple-http-server/config"
	"github.com/tangyang/simple-http-server/model"

	"fmt"
)

type UserDao struct {
}

func (u *UserDao) CreateUserSchema(conf *config.Config) error {
	c := NewPostgreConnector(conf)
	_, err := c.DB.Exec("CREATE TABLE users (id bigserial PRIMARY key , name CHARACTER VARYING)")
	return err
}

func (u *UserDao) AddUser(conf *config.Config, user *model.User) (bool, error) {
	c := NewPostgreConnector(conf)
	b, err := c.DB.Model(user).Where("name=?", user.Name).SelectOrCreate()
	// err := c.DB.Create(user)
	return b, err
}

func (u *UserDao) GetUserByName(conf *config.Config, name string) *model.User {
	c := NewPostgreConnector(conf)
	user := &model.User{}
	err := c.DB.Model(user).Where("name=?", name).Select()
	if err != nil {
		fmt.Sprintf("Fail to get user by name %s, error: %s\n", name, err.Error())
		return nil
	}
	return user
}

// func (u *UserDao) GetUserById(conf *config.Config, id int64) *model.User {
// 	c := NewPostgreConnector(conf)
// 	user := &model.User{}
// 	err := c.DB.Model(user).Where("id=?", id).Select()
// 	if err != nil {
// 		fmt.Sprintf("Fail to get user by id %d, error: %s\n", id, err.Error())
// 		return nil
// 	}
// 	return user
// }

func (u *UserDao) GetAllUsers(conf *config.Config) []model.User {
	c := NewPostgreConnector(conf)
	var users []model.User
	_, err := c.DB.Query(&users, `SELECT * FROM users`)
	if err != nil {
		fmt.Sprintf("Fail to get all users, error: %s\n", err.Error())
		return nil
	}
	return users
}
