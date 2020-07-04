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
    rtcpPLIInterval = time.Second * 3
)

var lock sync.RWMutex
var config webrtc.Configuration
var api *webrtc.API
var broadcasters map[string]*broadcaster

func init() {
    broadcasters = make(map[string]*broadcaster)
    mediaEngine := webrtc.MediaEngine{}
    mediaEngine.RegisterCodec(webrtc.NewRTPVP8Codec(webrtc.DefaultPayloadTypeVP8, 90000))
    mediaEngine.RegisterCodec(webrtc.NewRTPOpusCodec(webrtc.DefaultPayloadTypeOpus, 48000))
    api = webrtc.NewAPI(webrtc.WithMediaEngine(mediaEngine))
}

type writingTrack struct {
    track *webrtc.Track
    writer media.Writer
    filename string
}

type videoAudioTrack struct {
    video, audio *writingTrack
}

type broadcaster struct {
    Note interface{}
    identifier string
    tracks []*videoAudioTrack
}

func addBroadcaster(id string, quality int, identifier string) error {
    lock.Lock()
    defer lock.Unlock()
    if _, ok := broadcasters[id]; ok {
        return errors.New("")
    }
    broadcasters[id] = &broadcaster{tracks: make([]*videoAudioTrack, quality), identifier: identifier}
    return nil
}

func dropBroadcaster(id, identifier string) {
    if b, ok := broadcasters[id]; ok && b.identifier == identifier {
        delete(broadcasters, id)
    }
}

func setTrack(id string, quality int, video bool, identifier string, conn *webrtc.PeerConnection, remoteTrack *webrtc.Track) (*writingTrack, error) {
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
    if !ok || b.identifier != identifier {
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

func summarizeBroadcaster(id string) (res [][]string) {
    return
}

func newPeerConnectionWriter(id string, tracks map[string][]int, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
    identifier := "????????"
    peerConnection, err := api.NewPeerConnection(config)
    if err != nil {
        return nil, err
    }
    if _, err := peerConnection.AddTransceiverFromKind(webrtc.RTPCodecTypeVideo); err != nil {
        return nil, err
    }
    peerConnection.OnTrack(func(remoteTrack *webrtc.Track, receiver *webrtc.RTPReceiver) {
        pos, ok := tracks[remoteTrack.ID()]
        if !ok {
            return
        }
        localTrack, err := setTrack(id, pos[0], pos[1] == 0, identifier, peerConnection, remoteTrack)
        if err != nil {
            return
        }
        go func() {
            ticker := time.NewTicker(rtcpPLIInterval)
            for range ticker.C {
                if rtcpSendErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.PictureLossIndication{MediaSSRC: remoteTrack.SSRC()}}); rtcpSendErr != nil {
                    log.Println(rtcpSendErr)
                }
                if rtcpSendErr := peerConnection.WriteRTCP([]rtcp.Packet{&rtcp.ReceiverEstimatedMaximumBitrate{SenderSSRC: remoteTrack.SSRC(), Bitrate: 1e7}}); rtcpSendErr != nil {
                    log.Println(rtcpSendErr)
                }
            }
        }()
        for {
            rtp, err := remoteTrack.ReadRTP()
            if err != nil {
                panic(err)
            }
            // ErrClosedPipe means we don't have any subscribers, this is ok if no peers have connected yet
            if err = localTrack.track.WriteRTP(rtp); err != nil && err != io.ErrClosedPipe {
                panic(err)
            }
        }
    })
    peerConnection.OnICEConnectionStateChange(func(connectionState webrtc.ICEConnectionState) {
        if connectionState == webrtc.ICEConnectionStateFailed {
        }
    })
    if err := peerConnection.SetRemoteDescription(*offer); err != nil {
        dropBroadcaster(id, identifier)
        return nil, err
    }
    answer, err := peerConnection.CreateAnswer(nil)
    if err != nil {
        dropBroadcaster(id, identifier)
        return nil, err
    }
    if err := peerConnection.SetLocalDescription(answer); err != nil {
        dropBroadcaster(id, identifier)
        return nil, err
    }
    return &answer, nil
}

func newPeerConnectionReader(id string, quality int, offer *webrtc.SessionDescription) (*webrtc.SessionDescription, error) {
    peerConnection, err := api.NewPeerConnection(config)
    if err != nil {
        return nil, err
    }
    audioTrack, err := getTrack(id, quality, false)
    if err != nil {
        return nil, err
    }
    audioSender, err := peerConnection.AddTrack(audioTrack)
    if err != nil {
        return nil, err
    }
    go func() {
        for {
            pkg, err := audioSender.ReadRTCP()
            if err != nil {
                log.Println(err)
            } else {
                log.Println(pkg)
            }
        }
    }()
    videoTrack, err := getTrack(id, quality, true)
    if err != nil {
        return nil, err
    } else if videoTrack != nil {
        videoSender, err := peerConnection.AddTrack(videoTrack)
        if err != nil {
            return nil, err
        }
        go func() {
            for {
                pkg, err := videoSender.ReadRTCP()
                if err != nil {
                    log.Println(err)
                } else {
                    log.Println(pkg)
                }
            }
        }()
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
