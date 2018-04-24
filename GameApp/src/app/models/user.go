package models

import (
	. "app/consts"
	"gslib/baseModel"
)

type UserModel struct {
	baseModel.BaseModel
	Data *User
}
