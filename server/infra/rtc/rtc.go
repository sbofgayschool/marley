package rtc

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v3"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"

	"github.com/sbofgayschool/marley/server/utils"
)

const (
	rtcpInterval = time.Second * 3

	tempDir  = "temp/"
	mediaDir = "web/res/media/"

	videoClockRate  = 90000
	audioSampleRate = 48000

	ffmpegBin       = "ffmpeg.exe"
	ffmpegThread    = "-threads"
	ffmpegThreadNum = "5"
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
	writer   media.Writer
	filename string
}

type broadcaster struct {
	timestamp int64
	tracks    []*writingTrack
}

func addBroadcaster(id string, quality int, timestamp int64) error {
	lock.Lock()
	defer lock.Unlock()
	if _, ok := broadcasters[id]; ok {
		return errors.New("rtc broadcaster exists")
	}
	broadcasters[id] = &broadcaster{tracks: make([]*writingTrack, quality), timestamp: timestamp}
	return nil
}

func dropBroadcaster(id string, timestamp int64) {
	lock.Lock()
	defer lock.Unlock()
	if b, ok := broadcasters[id]; ok && b.timestamp == timestamp {
		delete(broadcasters, id)
	}
}

func setTrack(id string, quality int, timestamp int64, conn *webrtc.PeerConnection, remoteTrack *webrtc.Track) (*writingTrack, error) {
	lock.Lock()
	b, ok := broadcasters[id]
	lock.Unlock()
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
		name := tempDir + utils.RandomString() + ".ivf"
		writer, _ := utils.NewIVFWriter(name, qualityResolution[quality][0], qualityResolution[quality][1])
		b.tracks[quality] = &writingTrack{track: localTrack, writer: writer, filename: name}
		return b.tracks[quality], nil
	}
	name := tempDir + utils.RandomString() + ".ogg"
	writer, _ := oggwriter.New(name, audioSampleRate, 2)
	b.tracks[quality] = &writingTrack{track: localTrack, writer: writer, filename: name}
	return b.tracks[quality], nil
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

func NewPeerConnectionWriter(id string, timestamp int64, tracks []string, audioTimeStamp *int64, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
	stopped := false
	audioFile := make(chan string, 3)
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
			localTrack = nil
		}
		defer func() {
			if localTrack != nil && localTrack.writer != nil {
				if err = localTrack.writer.Close(); err != nil {
					log.Println(err)
				}
			}
			mergedFile := ""
			if quality == 0 {
				for i := 0; i < len(tracks)-1; i++ {
					if localTrack != nil && localTrack.writer != nil {
						audioFile <- localTrack.filename
					} else {
						audioFile <- ""
					}
				}
				close(audioFile)
				if localTrack != nil && localTrack.writer != nil {
					mergedFile = mediaDir + utils.RandomString() + ".ogg"
					if err := exec.Command(ffmpegBin, "-i", localTrack.filename, ffmpegThread, ffmpegThreadNum, mergedFile).Run(); err != nil {
						log.Println(err)
						mergedFile = ""
					}
				}
			} else {
				audio := <-audioFile
				if localTrack != nil && localTrack.writer != nil {
					mergedFile = mediaDir + utils.RandomString() + ".mp4"
					if audio == "" {
						if err := exec.Command(ffmpegBin, "-i", localTrack.filename, "-c:v", "h264", ffmpegThread, ffmpegThreadNum, "-s",
							fmt.Sprintf("%dx%d", qualityResolution[quality][0], qualityResolution[quality][1]), mergedFile).Run(); err != nil {
							log.Println(err)
							mergedFile = ""
						}
					} else {
						if err := exec.Command(ffmpegBin, "-i", localTrack.filename, "-c:v", "h264",
							"-i", audio, "-c:a", "aac", ffmpegThread, ffmpegThreadNum, "-s",
							fmt.Sprintf("%dx%d", qualityResolution[quality][0], qualityResolution[quality][1]), mergedFile).Run(); err != nil {
							log.Println(err)
							mergedFile = ""
						}
					}
				}
			}
			trackDoneCallback(id, timestamp, quality, mergedFile)
		}()
		if localTrack == nil {
			return
		}
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
		*audioTimeStamp = utils.UnixMillion()
		for {
			if stopped {
				break
			}
			rtp, err := remoteTrack.ReadRTP()
			if err != nil {
				// log.Println(err)
				continue
			}
			// ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
			if err = localTrack.track.WriteRTP(rtp); err != nil && err != io.ErrClosedPipe {
				log.Println(err)
			}
			if localTrack.writer != nil {
				if err = localTrack.writer.WriteRTP(rtp); err != nil {
					log.Println(err)
				}
			}
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
