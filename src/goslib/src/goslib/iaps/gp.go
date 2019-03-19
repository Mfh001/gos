package iaps

import (
	"context"
	"github.com/awa/go-iap/playstore"
	"goslib/gen_server"
	"goslib/logger"
	"io/ioutil"
)

type VerifyHandler func(productId string, success bool)

/*
   GenServer Callbacks
*/
type GPServer struct {
	bundleId string
	client   *playstore.Client
}

const GP_SERVER = "__GP_SERVEr__"

func StartGP(bundleId string, jsonKeyPath string) {
	gen_server.Start(GP_SERVER, new(GPServer), bundleId, jsonKeyPath)
}

func VerifyGP(productID string, purchaseToken string, handler VerifyHandler) {
	gen_server.Cast(GP_SERVER, "Verify", productID, purchaseToken, handler)
}

func (self *GPServer) Init(args []interface{}) (err error) {
	self.bundleId = args[0].(string)

	// You need to prepare a public key for your Android app's in app billing
	// at https://console.developers.google.com.
	// jsonKey.json
	jsonKeyPath := args[1].(string)
	jsonKey, err := ioutil.ReadFile(jsonKeyPath)
	if err != nil {
		logger.ERR("Start GP iap verify server failed: ", err)
		return err
	}
	self.client, err = playstore.New(jsonKey)
	if err != nil {
		logger.ERR("Start GP iap verify server failed: ", err)
		return err
	}

	return nil
}

func (self *GPServer) HandleCast(args []interface{}) {
	handle := args[0].(string)
	if handle == "Verify" {
		productID := args[1].(string)
		purchaseToken := args[2].(string)
		verifyHandler := args[3].(VerifyHandler)
		ctx := context.Background()
		resp, err := self.client.VerifyProduct(ctx, self.bundleId, productID, purchaseToken)
		if err != nil {
			logger.ERR("GP verify iap failed: ", err)
			verifyHandler(productID, false)
			return
		}
		verifyHandler(productID, resp.PurchaseState == 0)
	}
}

func (self *GPServer) HandleCall(args []interface{}) (interface{}, error) {
	return nil, nil
}

func (self *GPServer) Terminate(reason string) (err error) {
	return nil
}
