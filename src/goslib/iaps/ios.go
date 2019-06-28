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
	"github.com/awa/go-iap/appstore"
	"github.com/mafei198/gos/goslib/gen_server"
	"github.com/mafei198/gos/goslib/logger"
)

/*
   GenServer Callbacks
*/
type IOSServer struct {
	bundleId string
	client   *appstore.Client
}

const IOS_SERVER = "__IOS_SERVER__"

func StartIOS(bundleId string) {
	gen_server.Start(IOS_SERVER, new(IOSServer), bundleId)
}

func VerifyIOS(receipt string, handler VerifyHandler) {
	gen_server.Cast(IOS_SERVER, &IOSVerifyParams{receipt, handler})
}

func (self *IOSServer) Init(args []interface{}) (err error) {
	self.bundleId = args[0].(string)
	self.client = appstore.New()
	return nil
}

type IOSVerifyParams struct {
	receipt       string
	verifyHandler VerifyHandler
}

func (self *IOSServer) HandleCast(req *gen_server.Request) {
	switch params := req.Msg.(type) {
	case *IOSVerifyParams:
		req := appstore.IAPRequest{
			ReceiptData: params.receipt, // your receipt data encoded by base64
		}
		resp := &appstore.IAPResponse{}
		ctx := context.Background()
		err := self.client.Verify(ctx, req, resp)
		if err != nil {
			logger.ERR("Verify ios iap failed: ", err)
			return
		}
		if resp.Receipt.BundleID == self.bundleId {
			switch resp.Status {
			case 0:
				params.verifyHandler(resp.Receipt.InApp[0].ProductID, true)
			default:
				logger.ERR("IOS IAP verify status: ", resp.Status)
			}
		} else {
			params.verifyHandler(string(resp.Receipt.AppItemID), false)
		}
		break
	}
}

func (self *IOSServer) HandleCall(req *gen_server.Request) (interface{}, error) {
	return nil, nil
}

func (self *IOSServer) Terminate(reason string) (err error) {
	return nil
}
