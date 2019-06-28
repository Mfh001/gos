/*
The MIT License (MIT)

Copyright (c) 2018 SavinMax. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package iaps

import (
	"context"
	"github.com/awa/go-iap/playstore"
	"github.com/mafei198/gos/goslib/gen_server"
	"github.com/mafei198/gos/goslib/logger"
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
	gen_server.Cast(GP_SERVER, &VerifyParams{productID, purchaseToken, handler})
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

type VerifyParams struct {
	productID     string
	purchaseToken string
	verifyHandler VerifyHandler
}

func (self *GPServer) HandleCast(req *gen_server.Request) {
	switch params := req.Msg.(type) {
	case *VerifyParams:
		ctx := context.Background()
		resp, err := self.client.VerifyProduct(ctx, self.bundleId, params.productID, params.purchaseToken)
		if err != nil {
			logger.ERR("GP verify iap failed: ", err)
			params.verifyHandler(params.productID, false)
			return
		}
		params.verifyHandler(params.productID, resp.PurchaseState == 0)
		break
	}
}

func (self *GPServer) HandleCall(req *gen_server.Request) (interface{}, error) {
	return nil, nil
}

func (self *GPServer) Terminate(reason string) (err error) {
	return nil
}
