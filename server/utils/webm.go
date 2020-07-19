package utils

import (
	"log"
	"os"

	"github.com/at-wat/ebml-go/webm"

	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3/pkg/media/samplebuilder"
)

type WebmSaver struct {
	audioWriter, videoWriter       webm.BlockWriteCloser
	audioBuilder, videoBuilder     *samplebuilder.SampleBuilder
	audioTimestamp, videoTimestamp uint32
	width, height                  int
	Filename                       string
}

func NewWebmSaver(filename string, width int, height int) *WebmSaver {
	return &WebmSaver{
		audioBuilder: samplebuilder.New(10, &codecs.OpusPacket{}),
		videoBuilder: samplebuilder.New(10, &codecs.VP8Packet{}),
		Filename:     filename,
		width:        width,
		height:       height,
	}
}

func (s *WebmSaver) Close() {
	if s.audioWriter != nil {
		if err := s.audioWriter.Close(); err != nil {
			log.Println(err)
		}
	}
	if s.videoWriter != nil {
		if err := s.videoWriter.Close(); err != nil {
			log.Println(err)
		}
	}
}

func (s *WebmSaver) PushOpus(rtpPacket *rtp.Packet) bool {
	s.audioBuilder.Push(rtpPacket)
	for {
		sample := s.audioBuilder.Pop()
		if sample == nil {
			return true
		}
		if s.audioWriter == nil && s.height == 0 {
			s.InitWriter()
		}
		if s.audioWriter != nil {
			s.audioTimestamp += sample.Samples
			t := s.audioTimestamp / 48
			if _, err := s.audioWriter.Write(true, int64(t), sample.Data); err != nil {
				log.Println(err)
				return false
			}
		} else {
			return false
		}
	}
}
func (s *WebmSaver) PushVP8(rtpPacket *rtp.Packet) bool {
	s.videoBuilder.Push(rtpPacket)
	for {
		sample := s.videoBuilder.Pop()
		if sample == nil {
			return true
		}
		// Read VP8 header.
		videoKeyframe := (sample.Data[0]&0x1 == 0)
		if videoKeyframe {
			if s.videoWriter == nil || s.audioWriter == nil {
				s.InitWriter()
			}
		}
		if s.videoWriter != nil {
			s.videoTimestamp += sample.Samples
			t := s.videoTimestamp / 90
			if _, err := s.videoWriter.Write(videoKeyframe, int64(t), sample.Data); err != nil {
				log.Println(err)
				return false
			}
		} else {
			return false
		}
	}
}
func (s *WebmSaver) InitWriter() {
	w, err := os.OpenFile(s.Filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}
	ws, err := webm.NewSimpleBlockWriter(w,
		[]webm.TrackEntry{
			{
				Name:            "Audio",
				TrackNumber:     1,
				TrackUID:        12345,
				CodecID:         "A_OPUS",
				TrackType:       2,
				DefaultDuration: 20000000,
				Audio: &webm.Audio{
					SamplingFrequency: 48000.0,
					Channels:          2,
				},
			}, {
				Name:            "Video",
				TrackNumber:     2,
				TrackUID:        67890,
				CodecID:         "V_VP8",
				TrackType:       1,
				DefaultDuration: 33333333,
				Video: &webm.Video{
					PixelWidth:  uint64(s.width),
					PixelHeight: uint64(s.height),
				},
			},
		})
	if err != nil {
		panic(err)
	}
	s.audioWriter = ws[0]
	s.videoWriter = ws[1]
}
