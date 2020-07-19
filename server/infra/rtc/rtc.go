package rtc

import (
	"errors"
	"io"
	"log"
	"sync"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/sbofgayschool/marley/server/utils"
)

const (
	rtcpInterval = time.Second * 3

	tempDir  = "temp/"
	mediaDir = "web/res/media/"

	videoClockRate = 90000
	audioSampleRate = 48000
)

var qualityBitrate = []uint64{1e4, 5e5, 1e7, 5e8}
var qualityResolution = [][]int{{}, {256, 144}, {640, 360}, {1024, 576}}

var lock sync.RWMutex
var config webrtc.Configuration
var api *webrtc.API
var broadcasters map[string]*broadcaster

var connectionDoneCallback func(string, int64)

func SetConnectionDoneCallback(f func(string, int64)) {
	connectionDoneCallback = f
}

var trackDoneCallback func(string, int64, int, string)

func SetTrackDoneCallback(f func(string, int64, int, string)) {
	trackDoneCallback = f
}

func init() {
	broadcasters = make(map[string]*broadcaster)
	mediaEngine := webrtc.MediaEngine{}
	mediaEngine.RegisterCodec(webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, videoClockRate))
	mediaEngine.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, audioSampleRate))
	api = webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))
}

type writingTrack struct {
	track    *webrtc.Track
	writer   *utils.WebmSaver
	filename string
}

type broadcaster struct {
	timestamp int64
	unreadyTrack sync.WaitGroup
	tracks    []*writingTrack
}

func addBroadcaster(id string, quality int, timestamp int64) error {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := broadcasters[id]; ok {
		return errors.New("rtc broadcaster exists")
	}
	broadcasters[id] = &broadcaster{tracks: make([]*writingTrack, quality), timestamp: timestamp}
	broadcasters[id].unreadyTrack.Add(quality)
	return nil
}

func dropBroadcaster(id string, timestamp int64) {
	lock.Lock()
	defer lock.Unlock()
	if b, ok := broadcasters[id]; ok && b.timestamp == timestamp {
		delete(broadcasters, id)
	}
}

func setTrack(id string, quality int, timestamp int64, conn *webrtc.PeerConnection, remoteTrack *webrtc.Track) ([]*writingTrack, error) {
	lock.Lock()
	b, ok := broadcasters[id]
	lock.Unlock()
	if quality > 0 {
		defer b.unreadyTrack.Done()
	}
	if !ok || b.timestamp != timestamp {
		return nil, errors.New("broken peer connection")
	} else if len(b.tracks) <= quality {
		return nil, errors.New("incorrect quality parameter")
	}
	label := "video"
	if quality == 0 {
		label = "audio"
	}
	localTrack, err := conn.NewTrack(remoteTrack.PayloadType(), remoteTrack.SSRC(), remoteTrack.ID(), label)
	if err != nil {
		return nil, err
	}
	if quality > 0 {
		// writer := utils.NewWebmSaver(tempDir+remoteTrack.ID()+".webm", qualityResolution[quality][0], qualityResolution[quality][1])
		b.tracks[quality] = &writingTrack{track: localTrack, writer: nil, filename: remoteTrack.ID() + ".webm"}
		return []*writingTrack{b.tracks[quality]}, nil
	}
	// writer := utils.NewWebmSaver(tempDir+remoteTrack.ID()+".webm", 0, 0)
	b.tracks[quality] = &writingTrack{track: localTrack, writer: nil, filename: remoteTrack.ID() + ".webm"}
	b.unreadyTrack.Done()
	b.unreadyTrack.Wait()
	var res []*writingTrack
	for _, t := range b.tracks {
		if t != nil {
			res = append(res, t)
		}
	}
	return res, nil
}

func getTrack(id string, quality int) (*webrtc.Track, error) {
	lock.RLock()
	defer lock.RUnlock()
	if b, ok := broadcasters[id]; !ok {
		return nil, errors.New("no broadcaster")
	} else if quality >= len(b.tracks) {
		return nil, errors.New("no track")
	} else {
		if b.tracks[quality] == nil {
			return nil, errors.New("currently no track")
		}
		return b.tracks[quality].track, nil
	}
}

func NewPeerConnectionWriter(id string, timestamp int64, tracks []string, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	stopped := false
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}
	if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
		return nil, err
	}
	if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		return nil, err
	}
	log.Printf("tracks: %v\n", tracks)
	if err := addBroadcaster(id, len(tracks), timestamp); err != nil {
		return nil, err
	}
	peerConnection.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
		var quality int
		for i, tid := range tracks {
			if tid == remoteTrack.ID() {
				quality = i
				break
			}
		}
		log.Printf("track %v loaded: %v\n", remoteTrack.ID(), quality)
		localTrack, err := setTrack(id, quality, timestamp, peerConnection, remoteTrack)
		if err != nil {
			log.Println(err)
			return
		}
		defer func() {
			localTrack[0].writer.Close()
			trackDoneCallback(id, timestamp, quality, localTrack[0].filename)
		}()
		go func() {
			routinePacket := func() {
				if rtcpSendErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{
					MediaSSRC: remoteTrack.SSRC(),
				}}); rtcpSendErr != nil {
					log.Println(rtcpSendErr)
				}
				if rtcpSendErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.ReceiverEstimatedMaximumBitrate{
					SenderSSRC: remoteTrack.SSRC(),
					Bitrate:    qualityBitrate[quality],
				}}); rtcpSendErr != nil {
					log.Println(rtcpSendErr)
				}
			}
			routinePacket()
			ticker := time.NewTicker(rtcpInterval)
			for range ticker.C {
				if stopped {
					break
				}
				routinePacket()
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
			// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
			if err = localTrack[0].track.WriteRTP(rtp); err != nil && err != io.ErrClosedPipe {
				log.Println(err)
			}
			/*
			if quality > 0 {
				localTrack[0].writer.PushVP8(rtp)
			} else {
				for _, t := range localTrack {
					t.writer.PushOpus(rtp)
				}
			}
			*/
		}
	})
	peerConnection.OnConnectionStateChange(func(connectionState webrtc.PeerConnectionState) {
		log.Printf("writer %s: connection state changed %v\n", id, connectionState)
		if connectionState == webrtc.PeerConnectionStateFailed || connectionState == webrtc.PeerConnectionStateClosed {
			if !stopped {
				stopped = true
				_ = peerConnection.Close()
				dropBroadcaster(id, timestamp)
				connectionDoneCallback(id, timestamp)
			}
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

func NewPeerConnectionReader(id string, quality int, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	stopped := false
	peerConnection, err := api.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}
	audioTrack, err := getTrack(id, 0)
	if err != nil {
		return nil, err
	}
	if _, err := peerConnection.AddTrack(audioTrack); err != nil {
		return nil, err
	}
	if quality > 0 {
		videoTrack, err := getTrack(id, quality)
		if err != nil {
			return nil, err
		}
		if _, err := peerConnection.AddTrack(videoTrack); err != nil {
			return nil, err
		}
	}
	peerConnection.OnConnectionStateChange(func(connectionState webrtc.PeerConnectionState) {
		log.Printf("reader %s: connection state changed %v\n", id, connectionState)
		if connectionState == webrtc.PeerConnectionStateFailed || connectionState == webrtc.PeerConnectionStateClosed {
			if !stopped {
				stopped = true
				_ = peerConnection.Close()
			}
		}
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
