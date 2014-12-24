package register

import (
	. "app/consts"
	"app/controllers"
	"gslib"
	"gslib/routes"
)

func Load() {
	routes.Add(1, func(ctx interface{}, params interface{}) interface{} {
		instance := &controllers.EquipsController{Context: ctx.(*gslib.Player)}
		return instance.Load(params.(*EquipLoadParams))
	})
}
