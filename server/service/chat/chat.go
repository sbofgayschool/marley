package chat

import (
	"github.com/sbofgayschool/marley/server/infra/sock"
	"github.com/sbofgayschool/marley/server/service/common"
	"github.com/sbofgayschool/marley/server/service/user"
	"time"
)

const (
	Tag            = "chat"
	OperationField = "operation"
)

func init() {
	sock.RegisterHandler(Tag, sockHandler)
}

type Chat struct {
	Username    string
	MsgType     string
	Message     string
	ElapsedTime int64
}

var liveMessageCallback func(string, *Chat)

func SetLiveMessageCallback(f func(string, *Chat)) {
	liveMessageCallback = f
}

func chatMessage(id string, username string, msgType string, message string, elapsedTime int64) {
	ids := common.GetIdVodId(id)
	if len(ids) == 1 {
		liveMessageCallback(id, &Chat{Username: username, MsgType: msgType, Message: message, ElapsedTime: elapsedTime})
	} else {
		// TODO: Put the message directly into the database.
	}
}

func sockHandler(msg *sock.Message, broker chan *sock.Message) {
	content := msg.Content.(map[string]interface{})
	u := msg.Client.Info.(*user.SockUser)
	switch content[OperationField].(string) {
	case "NumQuery":
		broker <- &sock.Message{Client: msg.Client, Content: map[string]interface{}{"Num": common.GetCurrentAudience(msg.Client.Gid)}}
	case "Message":
		content["Username"] = u.Username
		elapsedTime := time.Now().Unix()
		if e, ok := content["ElapsedTime"]; ok {
			elapsedTime = e.(int64)
		}
		chatMessage(msg.Client.Gid, u.Username, content["MsgType"].(string), content["Message"].(string), elapsedTime)
		broker <- &sock.Message{Client: nil, Content: content}
	}
}
