package to

import (
	"github.com/tangyang/simple-http-server/model"
)

type UserTo struct {
	Id   int64
	Name string
	Type string
}

const (
	userType = "user"
)

func NewUserTo(user *model.User) *UserTo {
	return &UserTo{Id: user.Id, Name: user.Name, Type: userType}
}

func NewUserToArray(users []model.User) []UserTo {
	if users != nil {
		size := len(users)
		var result = []UserTo{}
		for i := 0; i < size; i++ {
			result = append(result, UserTo{Id: users[i].Id, Name: users[i].Name, Type: userType})
		}
		return result
	} else {
		return nil
	}
}
