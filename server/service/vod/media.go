package vod

import (
	"fmt"
	"github.com/sbofgayschool/marley/server/utils"
	"log"
	"os/exec"
)

const (
	tempDir   = "temp/"
	mediaDir  = "web/res/media/"
	ffmpegBin = "ffmpeg"
)

var qualityResolution = [][]int{{}, {256, 144}, {640, 360}, {1024, 576}}

func handleUploadedVideo(video int64, uploadedFile string, quality int) {
	log.Printf("%d %s %d transcode start\n", video, uploadedFile, quality)
	if quality > 0 && quality <= 3 {
		transcodeFile := utils.RandomString() + ".mp4"
		if err := exec.Command(ffmpegBin, "-i", tempDir+uploadedFile, "-s",
			fmt.Sprintf("%dx%d", qualityResolution[quality][0], qualityResolution[quality][1]),
			mediaDir+transcodeFile).Run(); err != nil {
			log.Println(err)
		} else if err := AddMedia(video, quality, transcodeFile); err != nil {
			log.Println(err)
		}
	} else if quality == 0 {
		transcodeFile := utils.RandomString() + ".mp3"
		if err := exec.Command(ffmpegBin, "-i", tempDir+uploadedFile, mediaDir+transcodeFile).Run(); err != nil {
			log.Println(err)
		} else if err := AddMedia(video, quality, transcodeFile); err != nil {
			log.Println(err)
		}
	}
	log.Printf("%d %s %d transcode done\n", video, uploadedFile, quality)
}
