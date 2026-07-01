package onvifmock

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/furrysalamander/onvif-mock-camera/internal/onvif"
	"github.com/furrysalamander/onvif-mock-camera/internal/rtsp"
	"github.com/furrysalamander/onvif-mock-camera/types"
)

type Camera struct {
	cfg         types.Config
	onvifServer *http.Server
	rtsp        *rtsp.Server
	mu          sync.Mutex
	running     bool
	stopCh      chan struct{}
}

func NewCamera(cfg types.Config) (*Camera, error) {
	if cfg.Source == nil {
		return nil, fmt.Errorf("VideoSource is required")
	}
	return &Camera{
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}, nil
}

func (c *Camera) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.running {
		return fmt.Errorf("camera already running")
	}

	cfg := c.cfg
	onvifAddr := fmt.Sprintf(":%d", cfg.OnvifPort)
	if cfg.OnvifPort == 0 {
		onvifAddr = fmt.Sprintf(":%d", types.DefaultOnvifPort)
	}

	frameCh, err := cfg.Source.Start()
	if err != nil {
		return fmt.Errorf("source start: %w", err)
	}

	rtsp, err := rtsp.NewServer(cfg)
	if err != nil {
		return fmt.Errorf("rtsp server: %w", err)
	}
	c.rtsp = rtsp

	if err := rtsp.StartServer(); err != nil {
		return fmt.Errorf("mediamtx: %w", err)
	}
	if err := rtsp.StartWithFrames(frameCh, cfg); err != nil {
		rtsp.Close()
		return fmt.Errorf("rtsp pipeline: %w", err)
	}

	c.onvifServer = onvif.StartServer(onvifAddr, cfg, cfg.Source)

	go onvif.RunDiscovery(cfg)

	c.running = true
	log.Printf("Camera started: ONVIF on %s, RTSP on :%d", onvifAddr, cfg.RtspPort)
	return nil
}

func (c *Camera) Stop() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.running {
		return nil
	}

	if c.onvifServer != nil {
		c.onvifServer.Shutdown(context.Background())
	}
	if c.rtsp != nil {
		c.rtsp.Close()
	}
	if c.cfg.Source != nil {
		c.cfg.Source.Stop()
	}

	c.running = false
	select {
	case c.stopCh <- struct{}{}:
	default:
	}
	log.Printf("Camera stopped")
	return nil
}

func (c *Camera) Run() error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
}

func (c *Camera) Wait() error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigCh:
		return c.Stop()
	case <-c.stopCh:
		return nil
	}
}
