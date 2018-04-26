package models

import (
	. "app/consts"
	"gslib/base_model"
)

type UserModel struct {
	base_model.BaseModel
	Data *User
}
