package iaps

import (
	"context"
	"github.com/awa/go-iap/appstore"
	"goslib/gen_server"
	"goslib/logger"
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
	receipt string
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
