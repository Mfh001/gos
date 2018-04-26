package base_model

import (
	"goslib/memstore"
	"gslib/player"
)

type BaseModel struct {
	TableName string
	Uuid      string
	Ctx       *player.Player
}

func (self *BaseModel) Save() {
	self.Ctx.Store.UpdateStatus(self.TableName, self.Uuid, memstore.STATUS_UPDATE)
}

func (self *BaseModel) Delete() {
	self.Ctx.Store.Del([]string{"models", self.TableName}, self.Uuid)
}
