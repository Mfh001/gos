package main

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gslib/player"
	"gslib/routes"
	"api"
	"app/models"
	"app/register"
	"goslib/memstore"
	"app/register/tables"
)

var _ = Describe("Game", func() {
	register.Load()
	memstore.InitDB()
	memstore.StartDBPersister()
	tables.RegisterTables(memstore.GetSharedDBInstance())

	It("should startup", func() {
		playerId := "fake_user_id"
		//player.CallPlayer(playerId, "handleWrap", func(ctx *player.Player) interface{} {
		//	return nil
		//})
		ctx := &player.Player{
			PlayerId: playerId,
		}
		ctx.Store = memstore.New(playerId, ctx)
		handler, _ := routes.Route("EquipLoadParams")
		params := &api.EquipLoadParams{
			PlayerID: playerId,
			EquipId: "fake_equip_id",
			HeroId: "fake_hero_id",
		}
		handler(ctx, params)
		user := models.FindUser(ctx, playerId)
		ctx.Store.Persist([]string{"models"})
		memstore.SyncPersistAll()
		Expect(user.Data.Level).Should(Equal(10))
	})
})
