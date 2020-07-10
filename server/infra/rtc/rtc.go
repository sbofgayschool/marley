package rtc

import (
	"errors"
	"io"
	"log"
	"sync"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
)

const (
	rtcpInterval = time.Second * 3
)

var qualityBitrate = [][]uint64{
	{1e7, 1e7},
	{1e7, 1e7},
	{1e7, 1e7},
}

var lock sync.RWMutex
var config webrtc.Configuration
var api *webrtc.API
var broadcasters map[string]*broadcaster

var connectionDoneCallback func(string, int64)

func SetConnectionDoneCallback(f func(string, int64)) {
	connectionDoneCallback = f
}

var trackDoneCallback func(string, int64, int, bool)

func SetTrackDoneCallback(f func(string, int64, int, bool)) {
	trackDoneCallback = f
}

func init() {
	broadcasters = make(map[string]*broadcaster)
	mediaEngine := webrtc.MediaEngine{}
	mediaEngine.RegisterCodec(webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000))
	mediaEngine.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))
	api = webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))
}

type writingTrack struct {
	track    *webrtc.Track
	writer   media.Writer
	filename string
}

type videoAudioTrack struct {
	video, audio *writingTrack
}

type broadcaster struct {
	timestamp int64
	tracks    []*videoAudioTrack
}

func addBroadcaster(id string, quality int, timestamp int64) error {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := broadcasters[id]; ok {
		return errors.New("")
	}
	broadcasters[id] = &broadcaster{tracks: make([]*videoAudioTrack, quality), timestamp: timestamp}
	return nil
}

func dropBroadcaster(id string, timestamp int64) {
	lock.Lock()
	defer lock.Unlock()
	if b, ok := broadcasters[id]; ok && b.timestamp == timestamp {
		delete(broadcasters, id)
	}
}

func setTrack(id string, quality int, video bool, timestamp int64, conn *webrtc.PeerConnection, remoteTrack *webrtc.Track) (*writingTrack, error) {
	lock.Lock()
	defer lock.Unlock()
	label := "audio"
	if video {
		label = "video"
	}
	localTrack, err := conn.NewTrack(remoteTrack.PayloadType(), remoteTrack.SSRC(), remoteTrack.ID(), label)
	if err != nil {
		return nil, err
	}
	b, ok := broadcasters[id]
	if !ok || b.timestamp != timestamp {
		return nil, errors.New("broken peer connection")
	} else if len(b.tracks) <= quality {
		return nil, errors.New("incorrect quality parameter")
	}
	if video {
		b.tracks[quality].video = &writingTrack{track: localTrack}
		return b.tracks[quality].video, nil
	}
	b.tracks[quality].audio = &writingTrack{track: localTrack}
	return b.tracks[quality].audio, nil
}

func getTrack(id string, quality int, video bool) (*webrtc.Track, error) {
	lock.RLock()
	defer lock.RUnlock()
	if b, ok := broadcasters[id]; !ok {
		return nil, errors.New("no broadcaster")
	} else if quality >= len(b.tracks) {
		return nil, errors.New("no track")
	} else {
		if video {
			return b.tracks[quality].video.track, nil
		}
		return b.tracks[quality].audio.track, nil
	}
}

func NewPeerConnectionWriter(id string, timestamp int64, tracks [][]string, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	stopped := false
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}
	if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		return nil, err
	}
	if err := addBroadcaster(id, len(tracks), timestamp); err != nil {
		return nil, err
	}
	peerConnection.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		var quality, video int
		for q, t := range tracks {
			for v, i := range t {
				if i == remoteTrack.ID() {
					quality = q
					video = v
				}
			}
		}
		localTrack, err := setTrack(id, quality, video == 1, timestamp, peerConnection, remoteTrack)
		if err != nil {
			log.Println(err)
			return
		}
		defer trackDoneCallback(id, timestamp, quality, video == 1)
		go func() {
			ticker := time.NewTicker(rtcpInterval)
			for range ticker.C {
				if stopped {
					break
				}
				if rtcpSendErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{
					MediaSSRC: remoteTrack.SSRC(),
				}}); rtcpSendErr != nil {
					log.Println(rtcpSendErr)
				}
				if rtcpSendErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.ReceiverEstimatedMaximumBitrate{
					SenderSSRC: remoteTrack.SSRC(),
					Bitrate:    qualityBitrate[quality][video],
				}}); rtcpSendErr != nil {
					log.Println(rtcpSendErr)
				}
			}
			ticker.Stop()
		}()
		for {
			if stopped {
				break
			}
			rtp, err := remoteTrack.ReadRTP()
			if err != nil {
				log.Println(err)
				continue
			}
			if err = localTrack.writer.WriteRTP(rtp); err != nil {
				log.Println(err)
			}
			// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
			if err = localTrack.track.WriteRTP(rtp); err != nil && err != io.ErrClosedPipe {
				log.Println(err)
			}
		}
	})
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
		if connectionState == webrtc.ICEConnectionStateFailed {
			stopped = true
			_ = peerConnection.Close()
			connectionDoneCallback(id, timestamp)
		}
	})
	if err := peerConnection.SetRemoteDescription(*offer); err != nil {
		dropBroadcaster(id, timestamp)
		return nil, err
	}
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		dropBroadcaster(id, timestamp)
		return nil, err
	}
	if err := peerConnection.SetLocalDescription(answer); err != nil {
		dropBroadcaster(id, timestamp)
		return nil, err
	}
	return &answer, nil
}

func NewPeerConnectionReader(id string, quality int, videoRequired bool, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}
	audioTrack, err := getTrack(id, quality, false)
	if err != nil {
		return nil, err
	}
	if _, err := peerConnection.AddTrack(audioTrack); err != nil {
		return nil, err
	}
	videoTrack, err := getTrack(id, quality, true)
	if err != nil {
		return nil, err
	} else if videoRequired && videoTrack != nil {
		if _, err := peerConnection.AddTrack(videoTrack); err != nil {
			return nil, err
		}
	}
	peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
	})
	if err := peerConnection.SetRemoteDescription(*offer); err != nil {
		return nil, err
	}
	answer, err := peerConnection.CreateAnswer(nil)
	if err != nil {
		return nil, err
	}
	if err := peerConnection.SetLocalDescription(answer); err != nil {
		return nil, err
	}
	return &answer, nil
}
