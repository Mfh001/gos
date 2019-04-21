package main

import (
	"app/custom_register/callbacks"
	"app/gen_register"
	"gen/api/pt"
	. "github.com/onsi/ginkgo"
	"goslib/api"
	"goslib/game_server"
	"goslib/logger"
	"goslib/player"
	"goslib/redisdb"
	"goslib/session_utils"
	"net"
	"time"
)

var _ = Describe("Game", func() {

	redisdb.StartClient()
	go game_server.Start("GS", func() {
		gen_register.RegisterRoutes()
		callbacks.RegisterSceneLoad()
	})

	accountId := "fakeAccountId"
	token := "fake_token"
	for {
		conn, err := net.Dial("tcp", "127.0.0.1:4000")
		if err != nil {
			logger.ERR("connect socket failed: ", err)
			time.Sleep(2 * time.Second)
			continue
		}
		session_utils.Create(&session_utils.Session{
			AccountId: accountId,
			GameAppId: player.CurrentGameAppId,
			Token:     token,
		})
		writer, err := api.Encode(pt.PT_SessionAuthParams, &pt.SessionAuthParams{
			AccountId: accountId,
			Token:     token,
		})
		data, _ := writer.GetSendData(0)
		conn.Write(data)

		writer, err = api.Encode(pt.PT_EquipLoadParams, &pt.EquipLoadParams{
			PlayerID: accountId,
			EquipId:  "fakeeid",
			HeroId:   "fakehero",
		})
		data, _ = writer.GetSendData(0)
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
