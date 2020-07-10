package live

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sbofgayschool/marley/server/infra/sock"
	"github.com/sbofgayschool/marley/server/service/user"
	"log"
	"time"

	"github.com/pion/webrtc/v3"

	"github.com/sbofgayschool/marley/server/infra/rtc"
)

const (
	Tag            = "liveMedia"
	OperationField = "operation"
)

type Broadcaster struct {
	Timestamp int64
	Qualities int
	AudioOnly bool
}

var broadcasters = make(map[string]*Broadcaster)

func connectionDoneCallback(id string, timestamp int64) {
	if b, ok := broadcasters[id]; ok && b.Timestamp == timestamp {
		// TODO: Put metadata into database.
		delete(broadcasters, id)
	}
}

func trackDoneCallback(id string, timestamp int64, quality int, video bool) {
	log.Printf("Track: %v %v %v %v is over\n", id, timestamp, quality, video)
	// TODO: Move the handled video to proper position, and put relative data into database
}

func init() {
	rtc.SetConnectionDoneCallback(connectionDoneCallback)
	rtc.SetTrackDoneCallback(trackDoneCallback)
	sock.RegisterHandler(Tag, sockHandler)
}

func Check(id string) *Broadcaster {
	if b, ok := broadcasters[id]; !ok {
		return nil
	} else {
		return b
	}
}

func join(id string, quality int, videoRequired bool, sdpType int, sdpString string) (*webrtc.SessionDescription, int64, error) {
	b, ok := broadcasters[id]
	if !ok {
		return nil, -1, errors.New("no broadcaster")
	}
	if ans, err := rtc.NewPeerConnectionReader(id, quality, videoRequired, &webrtc.SessionDescription{Type: webrtc.SDPType(sdpType), SDP: sdpString}); err != nil {
		return nil, -1, err
	} else {
		return ans, b.Timestamp, nil
	}
}

func add(id string, tracks [][]string, sdpType int, sdpString string) (*webrtc.SessionDescription, int64, error) {
	if _, ok := broadcasters[id]; ok {
		return nil, -1, errors.New("broadcaster exists")
	}
	t := time.Now().Unix()
	if ans, err := rtc.NewPeerConnectionWriter(id, t, tracks, &webrtc.SessionDescription{Type: webrtc.SDPType(sdpType), SDP: sdpString}); err != nil {
		return nil, -1, err
	} else {
		broadcasters[id] = &Broadcaster{Timestamp: t, Qualities: len(tracks), AudioOnly: len(tracks[0]) == 1}
		return ans, t, nil
	}
}

func sockHandler(msg *sock.Message, broker chan *sock.Message) {
	content := msg.Content.(map[string]interface{})
	u := msg.Client.Info.(user.SockUser)
	switch content["operation"] {
	case "check":
		broker <- &sock.Message{Client: msg.Client, Content: map[string]interface{}{
			sock.TagField:  Tag,
			OperationField: "check",
			"Broadcaster":  Check(msg.Client.Gid),
		}}
	case "add":
		offer := content["offer"].(map[string]interface{})
		var tracks [][]string

		answer, timestamp, err := add(msg.Client.Gid, tracks, offer["type"].(int), offer["sdp"].(string))
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
	case "join":
		if u.Teacher {
			offer := content["offer"].(map[string]interface{})
			answer, timestamp, err := join(msg.Client.Gid, content["Quality"].(int), content["VideoRequired"].(bool), offer["type"].(int), offer["sdp"].(string))
			errMessage := ""
			if err != nil {
				errMessage = err.Error()
			}
			broker <- &sock.Message{Client: msg.Client, Content: map[string]interface{}{
				sock.TagField:  Tag,
				OperationField: "join",
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
	}
}

func Upgrade(c *gin.Context) {
	id := c.Param("id")
	// uid := c.MustGet("uid").(int)
	if err := sock.NewClient(c, id, user.SockUser{Uid: 0, Username: "Anonymous User", Teacher: true}); err != nil {
	}
}
