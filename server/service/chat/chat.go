package chat

import (
	"github.com/sbofgayschool/marley/server/infra/sock"
	"github.com/sbofgayschool/marley/server/service/common"
	"github.com/sbofgayschool/marley/server/service/user"
	"github.com/sbofgayschool/marley/server/utils"
)

const (
	Tag            = "chat"
	OperationField = "Operation"
)

func init() {
	sock.RegisterHandler(Tag, sockHandler)
}

type Chat struct {
	Username    string
	MsgType     string
	Message     string
	Source      string
	ElapsedTime int64
	uid         int
}

var liveMessageCallback func(string, *Chat)

func SetLiveMessageCallback(f func(string, *Chat)) {
	liveMessageCallback = f
}

func chatMessage(id string, chat *Chat) {
	ids := common.GetIdVodId(id)
	if len(ids) == 1 {
		liveMessageCallback(id, chat)
	} else {
		// TODO: Put the message directly into the database.
	}
}

func sockHandler(msg *sock.Message, _ chan *sock.Message) (res []*sock.Message) {
	content := msg.Content.(map[string]interface{})
	u := msg.Client.Info.(*user.SockUser)
	switch content[OperationField].(string) {
	case "numQuery":
		res = append(res, &sock.Message{Client: msg.Client, Content: map[string]interface{}{
			sock.TagField:  Tag,
			OperationField: "numQuery",
			"Num":          msg.Client.GetPeerNum(),
		}})
	case "message":
		content["Username"] = u.Username
		elapsedTime := utils.UnixMillion()
		if e, ok := content["ElapsedTime"]; ok {
			elapsedTime = int64(e.(float64))
		}
		chat := &Chat{u.Username, content["MsgType"].(string), content["Message"].(string), content["Source"].(string), elapsedTime, u.Uid}
		chatMessage(msg.Client.Gid, chat)
		res = append(res, &sock.Message{Client: nil, Content: map[string]interface{}{
			sock.TagField:  Tag,
			OperationField: "message",
			"Username":     chat.Username,
			"MsgType":      chat.MsgType,
			"Message":      chat.Message,
			"Source":       chat.Source,
			"ElapsedTime":  chat.ElapsedTime,
		}})
	}
	return
}
