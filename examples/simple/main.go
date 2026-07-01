package main

import (
	"log"
	"time"

	"github.com/furrysalamander/onvif-mock-camera"
	"github.com/furrysalamander/onvif-mock-camera/doom"
	"github.com/furrysalamander/onvif-mock-camera/types"
)

func main() {
	src := doom.New(doom.Config{
		WADPath:     "doom1.wad",
		Skill:       2,
		WarpEpisode: 1,
		WarpMap:     1,
	})

	cam, err := onvifmock.NewCamera(types.Config{
		DeviceName: "My DOOM Camera",
		HostIP:     "127.0.0.1",
		OnvifPort:  8080,
		RtspPort:   8554,
		Source:     src,
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		time.Sleep(5 * time.Second)
		cam.Stop()
	}()

	if err := cam.Run(); err != nil {
		log.Fatal(err)
	}
}
