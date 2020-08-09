package chat

import (
	"github.com/sbofgayschool/marley/server/infra/sock"
	"github.com/sbofgayschool/marley/server/service/common"
	"github.com/sbofgayschool/marley/server/service/live"
	"github.com/sbofgayschool/marley/server/service/user"
	"github.com/sbofgayschool/marley/server/service/vod"
	"github.com/sbofgayschool/marley/server/utils"
	"log"
	"strconv"
)

const (
	Tag            = "chat"
	OperationField = "Operation"
)

func init() {
	sock.RegisterHandler(Tag, sockHandler)
}

func chatMessage(id string, chat *common.Chat) {
	cid, v := common.GetIdVodId(id)
	if v == "" {
		live.HandleMessage(cid, chat)
	} else {
		if vid, err := strconv.Atoi(v); err != nil {
			log.Println(err)
		} else if err := vod.AddChat(int64(vid), chat); err != nil {
			log.Println("failed to insert chat of Vod")
			log.Println(err)
		}
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
		chat := &common.Chat{u.Username, content["MsgType"].(string), content["Message"].(string), content["Source"].(string), elapsedTime, u.Uid}
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
