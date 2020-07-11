package live

import (
	"errors"

	"log"
	"time"

	"github.com/pion/webrtc/v3"

	"github.com/sbofgayschool/marley/server/infra/rtc"
	"github.com/sbofgayschool/marley/server/infra/sock"
	"github.com/sbofgayschool/marley/server/service/chat"
	"github.com/sbofgayschool/marley/server/service/user"
)

const (
	Tag            = "live"
	OperationField = "operation"
)

type Broadcaster struct {
	Timestamp  int64
	Qualities  int
	Pdf        string
	operations []*operation
	chats      []*chat.Chat
}

var broadcasters = make(map[string]*Broadcaster)

func connectionDoneCallback(id string, timestamp int64) {
	if b, ok := broadcasters[id]; ok && b.Timestamp == timestamp {
		delete(broadcasters, id)
		// TODO: Put chat message into database.
		// TODO: Put operations into database, if any.
	}
}

func trackDoneCallback(id string, timestamp int64, quality int, video bool, filename string) {
	log.Printf("Track: %v %v %v %v is over\n", id, timestamp, quality, video)
	// TODO: Move the handled video to proper position, and put relative data into database.
}

func liveMessageCallback(id string, c *chat.Chat) {
	if b, ok := broadcasters[id]; ok {
		c.ElapsedTime -= b.Timestamp
		b.chats = append(b.chats, c)
	}
}

func init() {
	rtc.SetConnectionDoneCallback(connectionDoneCallback)
	rtc.SetTrackDoneCallback(trackDoneCallback)
	chat.SetLiveMessageCallback(liveMessageCallback)
	sock.RegisterHandler(Tag, sockHandler)
}

func Check(id string) *Broadcaster {
	if b, ok := broadcasters[id]; !ok {
		return nil
	} else {
		return b
	}
}

func join(id string, quality int, videoRequired bool, sdpType int, sdpString string) (*webrtc.SessionDescription, *Broadcaster, error) {
	b, ok := broadcasters[id]
	if !ok {
		return nil, nil, errors.New("no broadcaster")
	}
	if ans, err := rtc.NewPeerConnectionReader(id, quality, videoRequired, &webrtc.SessionDescription{Type: webrtc.SDPType(sdpType), SDP: sdpString}); err != nil {
		return nil, nil, err
	} else {
		return ans, b, nil
	}
}

func add(id string, tracks [][]string, pdf string, sdpType int, sdpString string) (*webrtc.SessionDescription, int64, error) {
	if _, ok := broadcasters[id]; ok {
		return nil, -1, errors.New("broadcaster exists")
	}
	if len(tracks[0]) == 1 && len(pdf) == 0 {
		return nil, -1, errors.New("pdf file required")
	}
	t := time.Now().Unix()
	if ans, err := rtc.NewPeerConnectionWriter(id, t, tracks, &webrtc.SessionDescription{Type: webrtc.SDPType(sdpType), SDP: sdpString}); err != nil {
		return nil, -1, err
	} else {
		// TODO: Put metadata into database.
		broadcasters[id] = &Broadcaster{Timestamp: t, Qualities: len(tracks), Pdf: pdf}
		return ans, t, nil
	}
}

func sockHandler(msg *sock.Message, broker chan *sock.Message) {
	content := msg.Content.(map[string]interface{})
	u := msg.Client.Info.(*user.SockUser)
	switch content[OperationField] {
	case "check":
		broker <- &sock.Message{Client: msg.Client, Content: map[string]interface{}{
			sock.TagField:  Tag,
			OperationField: "check",
			"Broadcaster":  Check(msg.Client.Gid),
		}}
	case "add":
		if u.Teacher {
			offer := content["offer"].(map[string]interface{})
			var tracks [][]string
			for _, ts := range content["tracks"].([]interface{}) {
				var nt []string
				for _, t := range ts.([]interface{}) {
					nt = append(nt, t.(string))
				}
				tracks = append(tracks, nt)
			}
			answer, timestamp, err := add(msg.Client.Gid, tracks, offer["pdf"].(string), offer["type"].(int), offer["sdp"].(string))
			errMessage := ""
			if err != nil {
				errMessage = err.Error()
			}
			broker <- &sock.Message{Client: msg.Client, Content: map[string]interface{}{
				sock.TagField:  Tag,
				OperationField: "add",
				"Answer":       answer,
				"Timestamp":    timestamp,
				"Error":        errMessage,
			}}
		} else {
			broker <- &sock.Message{Client: msg.Client, Content: map[string]interface{}{
				sock.TagField:  Tag,
				OperationField: "join",
				"Answer":       "",
				"Timestamp":    -1,
				"Error":        "access denied",
			}}
		}
	case "join":
		offer := content["offer"].(map[string]interface{})
		if answer, b, err := join(msg.Client.Gid, content["Quality"].(int), content["VideoRequired"].(bool), offer["type"].(int), offer["sdp"].(string)); err != nil {
			broker <- &sock.Message{Client: msg.Client, Content: map[string]interface{}{
				sock.TagField:  Tag,
				OperationField: "join",
				"Answer":       "",
				"Timestamp":    -1,
				"Error":        err.Error(),
				"Operations":   nil,
				"Chats":        nil,
			}}
		} else {
			broker <- &sock.Message{Client: msg.Client, Content: map[string]interface{}{
				sock.TagField:  Tag,
				OperationField: "join",
				"Answer":       answer,
				"Timestamp":    b.Timestamp,
				"Error":        "",
				"Operations":   fetchOperations(b),
				"Chats":        b.chats,
			}}
		}
	case "opt":
		if u.Teacher {
			if err := addOperation(msg.Client.Gid, content["Opt"].(string), content["timestamp"].(int64)); err == nil {
				broker <- &sock.Message{Client: nil, Content: content}
			}
		}
	}
}
