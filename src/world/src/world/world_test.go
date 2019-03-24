package main

import (
	"gen/consts"
	"gen/models"
	"gen/register"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"goslib/memstore"
	"goslib/player"
)

var _ = Describe("World", func() {
	memstore.StartDB()
	memstore.StartDBPersister()
	register.RegisterTables(memstore.GetSharedDBInstance())

	It("should startup", func() {
		playerId := "fake_user_id"
		ctx := &player.Player{
			PlayerId: playerId,
		}
		ctx.Store = memstore.New(playerId, ctx)
		//handler, _ := routes.Route("EquipLoadParams")
		//params := &api.EquipLoadParams{
		//	PlayerID: playerId,
		//	EquipId: "fake_equip_id",
		//	HeroId: "fake_hero_id",
		//}
		//handler(ctx, params)

		user := models.CreateUser(ctx, &consts.User{
			Level:  1,
			Exp:    0,
			Online: true,
		})

		user = models.FindUser(ctx, user.GetUuid())
		user.Data.Level = 10
		user.Save()
		ctx.Store.Persist([]string{"models"})
		memstore.SyncPersistAll()
		Expect(user.Data.Level).Should(Equal(10))
	})
})
