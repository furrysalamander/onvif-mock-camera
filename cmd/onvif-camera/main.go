package main

import (
	"flag"
	"log"
	"os"

	"github.com/furrysalamander/onvif-mock-camera"
	"github.com/furrysalamander/onvif-mock-camera/doom"
	"github.com/furrysalamander/onvif-mock-camera/types"
)

var sourceFactories = map[string]func(cfg *types.Config, cmd string){}

func registerSource(name string, fn func(cfg *types.Config, cmd string)) {
	sourceFactories[name] = fn
}

func main() {
	wadPath := flag.String("wad", os.Getenv("ONVIF_WAD_PATH"), "Path to DOOM WAD file")
	skill := flag.Int("skill", 2, "Skill level (1-4)")
	warpEp := flag.Int("warp-episode", 1, "Starting episode")
	warpMap := flag.Int("warp-map", 1, "Starting map")
	waylandCmd := flag.String("wayland", os.Getenv("ONVIF_WAYLAND_CMD"), "Command to run as Wayland client")
	verbose := flag.Bool("verbose", false, "Verbose logging")

	flag.Parse()

	cfg := types.ConfigFromEnv()
	if *verbose {
		cfg.Verbose = true
	}

	switch {
	case *wadPath != "":
		cfg.Source = doom.New(doom.Config{
			WADPath:     *wadPath,
			Skill:       *skill,
			WarpEpisode: *warpEp,
			WarpMap:     *warpMap,
		})
		if cfg.DeviceName == types.DefaultDeviceName {
			cfg.DeviceName = "DOOM Camera"
		}
	case *waylandCmd != "":
		if fn, ok := sourceFactories["wayland"]; ok {
			fn(&cfg, *waylandCmd)
		} else {
			log.Fatal("wayland support not compiled in (build with -tags wayland)")
		}
	default:
		log.Fatal("no source specified: use --wad <path> or --wayland <cmd>")
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
