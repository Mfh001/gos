package notification

import (
	"fmt"
	"github.com/appleboy/go-fcm"
	"github.com/sideshow/apns2"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/payload"
	"goslib/gen_server"
	"goslib/logger"
	"log"
)

const IOS_SERVER = "__ios_notification_server__"
const FCM_SERVER = "__android_notification_server__"

const (
	CHANNEL_IOS = iota
	CHANNEL_ANDROID
)

/*
   GenServer Callbacks
*/
type Server struct {
	bundleId  string
	iosClient *apns2.Client
	fcmClient *fcm.Client
}

func StartIOS(bundleId string, iosCertP12Path string, iosP12Password string, isProduction bool) {
	gen_server.Start(IOS_SERVER, new(Server), IOS_SERVER, bundleId, iosCertP12Path, iosP12Password, isProduction)
}

func StartFCM(fcmAPIKey string) {
	gen_server.Start(FCM_SERVER, new(Server), FCM_SERVER, fcmAPIKey)
}

func Send(channel int, deviceToken string, content string) {
	switch channel {
	case CHANNEL_IOS:
		gen_server.Cast(IOS_SERVER, "SendIOS", deviceToken, content)
	case CHANNEL_ANDROID:
		gen_server.Cast(FCM_SERVER, "SendAndroid", deviceToken, content)
	}
}

func (self *Server) Init(args []interface{}) (err error) {
	category := args[0].(string)
	switch category {
	case IOS_SERVER:
		self.bundleId = args[1].(string)
		iosCertP12Path := args[2].(string)
		iosP12Password := args[3].(string)
		isProduction := args[4].(bool)

		cert, err := certificate.FromP12File(iosCertP12Path, iosP12Password)
		if err != nil {
			logger.ERR("Start ios push failed: ", err)
			return err
		}

		if isProduction {
			self.iosClient = apns2.NewClient(cert).Production()
		} else {
			self.iosClient = apns2.NewClient(cert).Development()
		}
		break
	case FCM_SERVER:
		fcmAPIKey := args[1].(string)
		self.fcmClient, err = fcm.NewClient(fcmAPIKey)
		if err != nil {
			logger.ERR("Start FCM failed: ", err)
			return err
		}
	}

	return nil
}

func (self *Server) HandleCast(args []interface{}) {
	handle := args[0].(string)
	deviceToken := args[1].(string)
	content := args[2].(string)
	if handle == "SendIOS" {
		notification := &apns2.Notification{}
		notification.DeviceToken = deviceToken
		notification.Topic = self.bundleId
		notification.Payload = payload.NewPayload().Alert(content).Badge(1)

		res, err := self.iosClient.Push(notification)

		if err != nil {
			logger.ERR("send ios push failed: ", err)
		}
		if res.Sent() {
			log.Println("Sent:", res.ApnsID)
		} else {
			fmt.Printf("Not Sent: %v %v %v\n", res.StatusCode, res.ApnsID, res.Reason)
		}
	} else if handle == "SendAndroid" {
		msg := &fcm.Message{
			To: deviceToken,
			Notification: &fcm.Notification{
				Body:  content,
				Badge: "1",
				Sound: "default",
			},
		}
		_, err := self.fcmClient.Send(msg)
		if err != nil {
			logger.ERR("send fcm push failed: ", err)
		}
	}
}

func (self *Server) HandleCall(args []interface{}) (interface{}, error) {
	handle := args[0].(string)
	if handle == "Dispatch" {
		// TODO
	}
	return nil, nil
}

func (self *Server) Terminate(reason string) (err error) {
	return nil
}
