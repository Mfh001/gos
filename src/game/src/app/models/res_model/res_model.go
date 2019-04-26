package res_model

import (
	"gen/api/pt"
	"gen/gd"
	"goslib/player"
	"goslib/utils"
)

func Enough(ctx *player.Player, res []gd.ResReq) bool {
	for _, req := range res {
		resource := ctx.Data.Resources[req.Category]
		if resource.Amount < int64(req.Amount) {
			return false
		}
	}
	return true
}

func Use(ctx *player.Player, res []gd.ResReq) []*pt.ResAmount {
	info := make([]*pt.ResAmount, 0)
	for _, req := range res {
		resource := ctx.Data.Resources[req.Category]
		resource.Amount = utils.MinInt64(resource.Amount, int64(req.Amount))
		info = append(info, &pt.ResAmount{
			Category: req.Category,
			Amount:   resource.Amount,
		})
	}
	return info
}

func AutoBuyFee(ctx *player.Player, res []gd.ResReq) int32 {
	var fee int32
	for _, req := range res {
		resource := ctx.Data.Resources[req.Category]
		if resource.Amount < int64(req.Amount) {
			fee += BuyFee(req.Category, int64(req.Amount)-resource.Amount)
		}
	}
	return fee
}

func BuyFee(category int32, amount int64) int32 {
	return 0
}

func TimeFee(seconds int32) int32 {
	return 0
}
