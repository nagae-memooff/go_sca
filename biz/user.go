package biz

import (
	"go_sca/models"
	"go_sca/public"
)

func FindUser(params interface{}) models.User {
	var user models.User
	public.Db.Where(params).First(&user)
	return user
}
