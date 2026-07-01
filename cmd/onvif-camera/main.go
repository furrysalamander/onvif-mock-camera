package main

import (
	"flag"
	"log"
	"os"

	"github.com/furrysalamander/onvif-mock-camera"
	"github.com/furrysalamander/onvif-mock-camera/doom"
)

func main() {
	wadPath := flag.String("wad", os.Getenv("ONVIF_WAD_PATH"), "Path to DOOM WAD file")
	skill := flag.Int("skill", 2, "Skill level (1-4)")
	warpEp := flag.Int("warp-episode", 1, "Starting episode (DOOM 1 only)")
	warpMap := flag.Int("warp-map", 1, "Starting map")
	verbose := flag.Bool("verbose", false, "Verbose logging")
	flag.Parse()

	cfg := onvifmock.ConfigFromEnv()

	if *verbose {
		cfg.Verbose = true
	}

	var src onvifmock.VideoSource

	switch {
	case *wadPath != "":
		src = doom.New(doom.Config{
			WADPath:     *wadPath,
			Skill:       *skill,
			WarpEpisode: *warpEp,
			WarpMap:     *warpMap,
		})
	default:
		log.Fatal("no video source specified: use --wad <path>")
	}

	cfg.Source = src

	if cfg.DeviceName == onvifmock.DefaultDeviceName {
		cfg.DeviceName = "DOOM Camera"
	}

	cam, err := onvifmock.NewCamera(cfg)
	if err != nil {
		log.Fatalf("create camera: %v", err)
	}

	log.Printf("Starting ONVIF camera...")
	log.Printf("  ONVIF: http://%s:%d/onvif/device_service", cfg.HostIP, cfg.OnvifPort)
	log.Printf("  RTSP:  rtsp://%s:%d/stream", cfg.HostIP, cfg.RtspPort)
	log.Printf("  Video: %dx%d @ %dfps", cfg.VideoWidth, cfg.VideoHeight, cfg.Framerate)

	if err := cam.Run(); err != nil {
		log.Fatalf("camera error: %v", err)
	}
}
