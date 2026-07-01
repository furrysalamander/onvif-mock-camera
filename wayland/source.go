//go:build wayland
package wayland

import (
	"fmt"
	"image"
	"log"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/mmulet/term.everything/wayland"
)

type Config struct {
	Command string
	Width   int
	Height  int
	FPS     int
}

type Source struct {
	cfg     Config
	frameCh chan *image.RGBA
	stopCh  chan struct{}
	mu      sync.Mutex
	running bool
}

func New(cfg Config) *Source {
	if cfg.Width == 0 {
		cfg.Width = 320
	}
	if cfg.Height == 0 {
		cfg.Height = 200
	}
	if cfg.FPS == 0 {
		cfg.FPS = 35
	}
	return &Source{
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}
}

type displayArgs struct {
	name string
}

func (a *displayArgs) WaylandDisplayName() string {
	return a.name
}

func (s *Source) Start() (<-chan *image.RGBA, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return nil, fmt.Errorf("already running")
	}

	args := &displayArgs{name: ""}
	listener, err := wayland.MakeSocketListener(args)
	if err != nil {
		return nil, fmt.Errorf("wayland socket: %w", err)
	}

	go listener.MainLoopThenClose()

	desktop := wayland.MakeDesktop(
		wayland.Size{Width: uint32(s.cfg.Width), Height: uint32(s.cfg.Height)},
		false,
		nil,
	)

	var (
		clients   []*wayland.Client
		clientsMu sync.Mutex
		frameTick = make(chan struct{}, 1)
	)

	go func() {
		for conn := range listener.OnConnection {
			client := wayland.MakeClient(conn)
			clientsMu.Lock()
			clients = append(clients, client)
			clientsMu.Unlock()
			go client.MainLoop()
			go func(c *wayland.Client) {
				for range c.FrameDrawRequests {
					select {
					case frameTick <- struct{}{}:
					default:
					}
				}
			}(client)
		}
	}()

	if s.cfg.Command != "" {
		cmd := exec.Command("sh", "-c", s.cfg.Command)
		cmd.Env = append(os.Environ(),
			"WAYLAND_DISPLAY="+listener.WaylandDisplayName,
			"XDG_SESSION_TYPE=wayland",
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Start(); err != nil {
			log.Printf("wayland: command start: %v", err)
		}
	}

	s.frameCh = make(chan *image.RGBA, 4)
	s.running = true

	go s.captureLoop(desktop, &clients, &clientsMu, frameTick)

	return s.frameCh, nil
}

func (s *Source) captureLoop(desktop *wayland.Desktop, clients *[]*wayland.Client, mu *sync.Mutex, frameTick chan struct{}) {
	defer close(s.frameCh)

	ticker := time.NewTicker(time.Second / time.Duration(s.cfg.FPS))
	defer ticker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-frameTick:
		case <-ticker.C:
		}

		mu.Lock()
		desktop.Clear()
		desktop.DrawClients(*clients)
		mu.Unlock()

		frame := image.NewRGBA(image.Rect(0, 0, s.cfg.Width, s.cfg.Height))
		copy(frame.Pix, desktop.Buffer[:s.cfg.Width*s.cfg.Height*4])

		select {
		case s.frameCh <- frame:
		case <-s.stopCh:
			return
		}
	}
}

func (s *Source) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.running {
		return nil
	}
	select {
	case s.stopCh <- struct{}{}:
	default:
	}
	s.running = false
	return nil
}

func (s *Source) FrameSize() (int, int) {
	return s.cfg.Width, s.cfg.Height
}

func (s *Source) ContinuousMove(pan, tilt, zoom float64) error {
	return nil
}

func (s *Source) RelativeMove(pan, tilt, zoom float64) error {
	return nil
}

func (s *Source) AbsoluteMove(pan, tilt, zoom float64) error {
	return nil
}

func (s *Source) StopPTZ(panTilt, zoom bool) error {
	return nil
}
