package main

import (
	"app/custom_register"
	"app/custom_register/callbacks"
	"app/gen_register"
	"gen/api/pt"
	. "github.com/onsi/ginkgo"
	"goslib/api"
	"goslib/game_server"
	"goslib/logger"
	"net"
	"time"
)

var _ = Describe("Game", func() {

	go game_server.Start(func() {
		gen_register.RegisterRoutes()
		custom_register.RegisterCustomDataLoader()
		callbacks.RegisterSceneLoad()
	})

	for {
		conn, err := net.Dial("tcp", "127.0.0.1:4000")
		if err != nil {
			logger.ERR("connect socket failed: ", err)
			time.Sleep(1 * time.Second)
			continue
		}
		writer, err := api.Encode(pt.PT_SessionAuthParams, &pt.SessionAuthParams{
			AccountId: "fakeAccountId",
			Token:     "fakeeid",
		})
		data, _ := writer.GetSendData()
		conn.Write(data)

		writer, err = api.Encode(pt.PT_EquipLoadParams, &pt.EquipLoadParams{
			PlayerID: "fakeAccountId",
			EquipId:  "fakeeid",
			HeroId:   "fakehero",
		})
		data, _ = writer.GetSendData()
		conn.Write(data)

		time.Sleep(1000 * time.Second)
		break
	}

	//custom_register.RegisterCustomDataLoader()
	//memstore.StartDB()
	//memstore.StartDBPersister()
	//register.RegisterTables(memstore.GetSharedDBInstance())
	//
	//It("should startup", func() {
	//	playerId := "fake_user_id"
	//	ctx := &player.Player{
	//		PlayerId: playerId,
	//	}
	//	ctx.Store = memstore.New(playerId, ctx)
	//	//handler, _ := routes.Route("EquipLoadParams")
	//	//params := &api.EquipLoadParams{
	//	//	PlayerID: playerId,
	//	//	EquipId: "fake_equip_id",
	//	//	HeroId: "fake_hero_id",
	//	//}
	//	//handler(ctx, params)
	//
	//	user := models.CreateUser(ctx, &consts.User{
	//		Level:  1,
	//		Exp:    0,
	//		Online: true,
	//	})
	//
	//	user = models.FindUser(ctx, user.GetUuid())
	//	user.Data.Level = 10
	//	user.Save()
	//	ctx.Store.Persist([]string{"models"})
	//	memstore.SyncPersistAll()
	//	Expect(user.Data.Level).Should(Equal(10))
	//})
})
