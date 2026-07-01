package doom

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/AndreRenaud/gore"
)

const keyEnter = 13

type Config struct {
	WADPath     string
	PWADPath    string
	Skill       int
	WarpEpisode int
	WarpMap     int
}

type Source struct {
	cfg      Config
	frameCh  chan *image.RGBA
	eventCh  chan gore.DoomEvent
	stopCh   chan struct{}
	mu       sync.Mutex
	keysHeld map[uint8]bool
	running  bool
}

func New(cfg Config) *Source {
	return &Source{
		cfg:      cfg,
		keysHeld: make(map[uint8]bool),
		stopCh:   make(chan struct{}),
	}
}

func (s *Source) Start() (<-chan *image.RGBA, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.running {
		return nil, fmt.Errorf("already running")
	}

	s.frameCh = make(chan *image.RGBA, 4)
	s.eventCh = make(chan gore.DoomEvent, 32)

	go func() {
		defer close(s.frameCh)
		defer func() { s.running = false }()

		frontend := &doomFrontend{
			source: s,
		}

		wadDir, wadFile := filepath.Split(s.cfg.WADPath)
		if wadDir == "" {
			wadDir = "."
		}
		wadDir = filepath.Clean(wadDir)
		gore.SetVirtualFileSystem(os.DirFS(wadDir))

		args := []string{"-iwad", wadFile}
		if s.cfg.PWADPath != "" {
			pwadFile := filepath.Base(s.cfg.PWADPath)
			args = append(args, "-file", pwadFile)
		}
		if s.cfg.WarpEpisode > 0 && s.cfg.WarpMap > 0 {
			args = append(args, "-warp", fmt.Sprintf("%d", s.cfg.WarpEpisode), fmt.Sprintf("%d", s.cfg.WarpMap))
		}
		if s.cfg.Skill > 0 {
			skillMap := []string{"", "1", "2", "3", "4"}
			if s.cfg.Skill < len(skillMap) {
				args = append(args, "-skill", skillMap[s.cfg.Skill])
			}
		}

		s.running = true

		go func() {
			gore.EnableQuitting(false)
			gore.Run(frontend, args)
		}()

		// auto-start: skip title/demo screens after engine boots
		go func() {
			time.Sleep(2 * time.Second)
			for i := 0; i < 5; i++ {
				select {
				case s.eventCh <- gore.DoomEvent{Type: gore.Ev_keydown, Key: keyEnter}:
				default:
				}
				time.Sleep(200 * time.Millisecond)
				select {
				case s.eventCh <- gore.DoomEvent{Type: gore.Ev_keyup, Key: keyEnter}:
				default:
				}
				time.Sleep(100 * time.Millisecond)
			}
		}()

		<-s.stopCh
		gore.Stop()
	}()

	return s.frameCh, nil
}

func (s *Source) Stop() error {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil
	}
	s.mu.Unlock()

	select {
	case s.stopCh <- struct{}{}:
	default:
	}
	return nil
}

func (s *Source) FrameSize() (int, int) {
	return 320, 200
}

func (s *Source) ContinuousMove(pan, tilt, zoom float64) error {
	s.injectKeysForVelocity(pan, tilt, zoom)
	return nil
}

func (s *Source) RelativeMove(pan, tilt, zoom float64) error {
	key := uint8(0)
	switch {
	case pan > 0.1:
		key = gore.KEY_RIGHTARROW1
	case pan < -0.1:
		key = gore.KEY_LEFTARROW1
	case tilt > 0.1:
		key = gore.KEY_UPARROW1
	case tilt < -0.1:
		key = gore.KEY_DOWNARROW1
	case zoom > 0.1:
		key = gore.KEY_FIRE1
	case zoom < -0.1:
		key = gore.KEY_USE1
	}
	if key != 0 {
		s.sendKey(key, true)
		time.Sleep(50 * time.Millisecond)
		s.sendKey(key, false)
	}
	return nil
}

func (s *Source) AbsoluteMove(pan, tilt, zoom float64) error {
	return nil
}

func (s *Source) StopPTZ(panTilt, zoom bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for key := range s.keysHeld {
		s.sendKeyUnlocked(key, false)
	}
	s.keysHeld = make(map[uint8]bool)
	return nil
}

func (s *Source) injectKeysForVelocity(pan, tilt, zoom float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	wantKeys := make(map[uint8]bool)
	velToKey := map[uint8]float64{
		gore.KEY_RIGHTARROW1: pan,
		gore.KEY_LEFTARROW1:  -pan,
		gore.KEY_UPARROW1:    tilt,
		gore.KEY_DOWNARROW1:  -tilt,
		gore.KEY_FIRE1:       zoom,
		gore.KEY_USE1:        -zoom,
	}

	for key, vel := range velToKey {
		if vel > 0.05 {
			wantKeys[key] = true
		}
	}

	for key := range s.keysHeld {
		if !wantKeys[key] {
			s.sendKeyUnlocked(key, false)
		}
	}
	for key := range wantKeys {
		if !s.keysHeld[key] {
			s.sendKeyUnlocked(key, true)
		}
	}
	s.keysHeld = wantKeys
}

func (s *Source) sendKey(key uint8, pressed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sendKeyUnlocked(key, pressed)
}

func (s *Source) sendKeyUnlocked(key uint8, pressed bool) {
	typ := gore.Ev_keydown
	if !pressed {
		typ = gore.Ev_keyup
	}
	select {
	case s.eventCh <- gore.DoomEvent{Type: typ, Key: key}:
	default:
	}
}

type doomFrontend struct {
	source *Source
}

func (f *doomFrontend) DrawFrame(img *image.RGBA) {
	// gore reuses the same backing buffer — copy to avoid races with the encoder
	copied := image.NewRGBA(img.Bounds())
	copy(copied.Pix, img.Pix)
	select {
	case f.source.frameCh <- copied:
	default:
	}
}

func (f *doomFrontend) GetEvent(event *gore.DoomEvent) bool {
	select {
	case ev := <-f.source.eventCh:
		*event = ev
		return true
	default:
		return false
	}
}

func (f *doomFrontend) SetTitle(title string) {}
func (f *doomFrontend) CacheSound(name string, data []byte) {}
func (f *doomFrontend) PlaySound(name string, channel, volume, sep int) {}
