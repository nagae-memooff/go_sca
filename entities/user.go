package entities

import (
	"go_sca/models"
)

type User struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Actived bool   `json:"actived"`
}

func NewUser(user models.User) User {
	_user := User{
		ID:      user.ID,
		Name:    user.Name,
		Actived: user.Actived,
	}

	return _user
}

func NewUsers(users []models.User) []User {
	list := make([]User, 0, len(users))
	for _, u := range users {
		list = append(list, NewUser(u))
	}
	return list
}
