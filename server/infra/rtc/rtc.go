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
	{1e4, 5e5},
	{2e4, 1e7},
	{3e4, 5e8},
}

var lock sync.RWMutex
var config webrtc.Configuration
var api *webrtc.API
var broadcasters map[string]*broadcaster

var connectionDoneCallback func(string, int64)

func SetConnectionDoneCallback(f func(string, int64)) {
	connectionDoneCallback = f
}

var trackDoneCallback func(string, int64, int, bool, string)

func SetTrackDoneCallback(f func(string, int64, int, bool, string)) {
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
		return errors.New("rtc broadcaster exists")
	}
	broadcasters[id] = &broadcaster{tracks: make([]*videoAudioTrack, quality), timestamp: timestamp}
	for i := 0; i < quality; i++ {
		broadcasters[id].tracks[i] = new(videoAudioTrack)
	}
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
	b, ok := broadcasters[id]
	lock.Unlock()
	if !ok || b.timestamp != timestamp {
		return nil, errors.New("broken peer connection")
	} else if len(b.tracks) <= quality {
		return nil, errors.New("incorrect quality parameter")
	}
	label := "audio"
	if video {
		label = "video"
	}
	localTrack, err := conn.NewTrack(remoteTrack.PayloadType(), remoteTrack.SSRC(), remoteTrack.ID(), label)
	if err != nil {
		return nil, err
	}
	if video {
		// TODO: Add information for writer and filename, and start the writer
		b.tracks[quality].video = &writingTrack{track: localTrack}
		return b.tracks[quality].video, nil
	}
	// TODO: Add information for writer and filename, and start the writer
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
			if b.tracks[quality].video == nil {
				return nil, errors.New("currently no track")
			}
			return b.tracks[quality].video.track, nil
		}
		if b.tracks[quality].audio == nil {
			return nil, errors.New("currently no track")
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
	if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeAudio); err != nil {
		return nil, err
	}
	log.Printf("tracks: %v\n", tracks)
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
		log.Printf("track %v loaded: %v %v\n", remoteTrack.ID(), quality, video)
		localTrack, err := setTrack(id, quality, video == 1, timestamp, peerConnection, remoteTrack)
		if err != nil {
			log.Println(err)
			return
		}
		defer func() {
			// TODO: Stop the writer and finalized the file.
			trackDoneCallback(id, timestamp, quality, video == 1, localTrack.filename)
		}()
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
					// TODO: Figure out the bitrate.
					Bitrate:    qualityBitrate[quality][video],
					// Bitrate: 1e8,
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
				// log.Println(err)
				continue
			}
			/*
			if err = localTrack.writer.WriteRTP(rtp); err != nil {
				log.Println(err)
			}
			*/
			// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
			if err = localTrack.track.WriteRTP(rtp); err != nil && err != io.ErrClosedPipe {
				log.Println(err)
			}
			// TODO: Write the file
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

func NewPeerConnectionReader(id string, quality int, videoRequired bool, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	stopped := false
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
