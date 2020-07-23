package live

import (
	"errors"
	"github.com/sbofgayschool/marley/server/utils"
	"sync"

	"log"

	"github.com/pion/webrtc/v3"

	"github.com/sbofgayschool/marley/server/infra/rtc"
	"github.com/sbofgayschool/marley/server/infra/sock"
	"github.com/sbofgayschool/marley/server/service/chat"
	"github.com/sbofgayschool/marley/server/service/user"
)

const (
	Tag            = "live"
	OperationField = "Operation"
)

type Broadcaster struct {
	Timestamp  int64
	audioTimestamp *int64
	Qualities  int
	Pdf        string
	operations []*operation
	chats      []*chat.Chat
}

var broadcasters = make(map[string]*Broadcaster)
var lock sync.RWMutex

func connectionDoneCallback(id string, timestamp int64) {
	log.Printf("connection: %v %v is over\n", id, timestamp)
	lock.Lock()
	if b, ok := broadcasters[id]; ok && b.Timestamp == timestamp {
		delete(broadcasters, id)
		lock.Unlock()
		// TODO: Put chat message into database.
		// TODO: Put operations into database, if any.
	} else {
		lock.Unlock()
	}
}

func trackDoneCallback(id string, timestamp int64, quality int, filename string) {
	log.Printf("track: %v %v %v %v is over\n", id, timestamp, quality, filename)
	// TODO: Move the handled video to proper position, and put relative data into database.
}

func liveMessageCallback(id string, c *chat.Chat) {
	lock.RLock()
	b, ok := broadcasters[id]
	lock.RUnlock()
	if ok {
		b.chats = append(b.chats, c)
	}
}

func init() {
	rtc.SetConnectionDoneCallback(connectionDoneCallback)
	rtc.SetTrackDoneCallback(trackDoneCallback)
	chat.SetLiveMessageCallback(liveMessageCallback)
	sock.RegisterHandler(Tag, sockHandler)
}

func check(id string) *Broadcaster {
	lock.RLock()
	defer lock.RUnlock()
	if b, ok := broadcasters[id]; !ok {
		return nil
	} else {
		return b
	}
}

func join(id string, quality int, sdpString string) (*webrtc.SessionDescription, int64, error) {
	var timestamp int64 = -1
	lock.RLock()
	if b, ok := broadcasters[id]; ok {
		timestamp = b.Timestamp
	}
	lock.RUnlock()
	if timestamp == -1 {
		return nil, -1, errors.New("no broadcaster")
	}
	if ans, err := rtc.NewPeerConnectionReader(id, quality, &webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: sdpString}); err != nil {
		return nil, -1, err
	} else {
		return ans, timestamp, nil
	}
}

func add(id string, tracks []string, pdf string, sdpString string) (*webrtc.SessionDescription, int64, error) {
	lock.RLock()
	_, ok := broadcasters[id]
	lock.RUnlock()
	if ok {
		return nil, -1, errors.New("broadcaster exists")
	}
	t := utils.UnixMillion()
	var audioTimestamp int64
	if ans, err := rtc.NewPeerConnectionWriter(id, t, tracks, &audioTimestamp, &webrtc.SessionDescription{Type: webrtc.SDPTypeOffer, SDP: sdpString}); err != nil {
		return nil, -1, err
	} else {
		// TODO: Put metadata into database.
		lock.Lock()
		defer lock.Unlock()
		broadcasters[id] = &Broadcaster{Timestamp: t, Qualities: len(tracks), Pdf: pdf, audioTimestamp: &audioTimestamp}
		return ans, t, nil
	}
}

func sockHandler(msg *sock.Message, broker chan *sock.Message) (res []*sock.Message) {
	content := msg.Content.(map[string]interface{})
	u := msg.Client.Info.(*user.SockUser)
	switch content[OperationField] {
	case "check":
		res = append(res, &sock.Message{Client: msg.Client, Content: map[string]interface{}{
			sock.TagField:  Tag,
			OperationField: "check",
			"Broadcaster":  check(msg.Client.Gid),
		}})
	case "add":
		if u.Teacher {
			go func() {
				offer := content["Offer"].(map[string]interface{})
				var tracks []string
				for _, ts := range content["Tracks"].([]interface{}) {
					tracks = append(tracks, ts.(string))
				}
				answer, timestamp, err := add(msg.Client.Gid, tracks, content["Pdf"].(string), offer["sdp"].(string))
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
			}()
		} else {
			res = append(res, &sock.Message{Client: msg.Client, Content: map[string]interface{}{
				sock.TagField:  Tag,
				OperationField: "join",
				"Answer":       "",
				"Timestamp":    -1,
				"Error":        "access denied",
			}})
		}
	case "join":
		go func() {
			offer := content["Offer"].(map[string]interface{})
			if answer, t, err := join(msg.Client.Gid, int(content["Quality"].(float64)), offer["sdp"].(string)); err != nil {
				broker <- &sock.Message{Client: msg.Client, Content: map[string]interface{}{
					sock.TagField:  Tag,
					OperationField: "join",
					"Answer":       "",
					"Timestamp":    -1,
					"Error":        err.Error(),
				}}
			} else {
				broker <- &sock.Message{Client: msg.Client, Content: map[string]interface{}{
					sock.TagField:  Tag,
					OperationField: "join",
					"Answer":       answer,
					"Timestamp":    t,
					"Error":        "",
				}}
			}
		}()
	case "opt":
		if u.Teacher {
			if err := addOperation(msg.Client.Gid, int64(content["Timestamp"].(float64)), content["Opt"].(string)); err == nil {
				res = append(res, &sock.Message{Client: nil, Content: content})
			}
		}
	case "fetch":
		b := check(msg.Client.Gid)
		if b == nil {
			res = append(res, &sock.Message{Client: msg.Client, Content: map[string]interface{}{
				sock.TagField:  Tag,
				OperationField: "fetch",
				"Timestamp":    -1,
				"Operations":   nil,
				"Chats":        nil,
				"Error":        "no broadcaster",
			}})
		} else {
			res = append(res, &sock.Message{Client: msg.Client, Content: map[string]interface{}{
				sock.TagField:  Tag,
				OperationField: "fetch",
				"Timestamp":    b.Timestamp,
				"Operations":   fetchOperations(b),
				"Chats":        b.chats,
				"Error":        "",
			}})
		}
	}
	return
}
