package onvifmock

import (
	"context"
	"fmt"
	"image"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type rtspServer struct {
	cfg          Config
	mediamtxCmd  *exec.Cmd
	ffmpegCmd    *exec.Cmd
	cancel       context.CancelFunc
	mu           sync.Mutex
	started      bool
	ffmpegPath   string
	mediamtxPath string
}

func newRTSPServer(cfg Config) (*rtspServer, error) {
	ffmpeg := cfg.FFmpegPath
	if ffmpeg == "" {
		ffmpeg = "ffmpeg"
	}
	mediamtx := cfg.MediamtxPath
	if mediamtx == "" {
		mediamtx = "mediamtx"
	}
	return &rtspServer{
		cfg:          cfg,
		ffmpegPath:   ffmpeg,
		mediamtxPath: mediamtx,
	}, nil
}

func (rs *rtspServer) startServer() error {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	rtspPort := rs.cfg.RtspPort
	if rtspPort == 0 {
		rtspPort = DefaultRtspPort
	}

	configYAML := fmt.Sprintf(`rtspAddress: :%d
rtmp: no
hls: no
webrtc: no
logLevel: warn
paths:
  all:
    source: publisher
`, rtspPort)

	configDir, err := os.MkdirTemp("", "mediamtx-")
	if err != nil {
		return fmt.Errorf("temp dir: %w", err)
	}
	configPath := filepath.Join(configDir, "mediamtx.yml")
	if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	rs.cancel = cancel

	rs.mediamtxCmd = exec.CommandContext(ctx, rs.mediamtxPath, configPath)

	if err := rs.mediamtxCmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("mediamtx start: %w", err)
	}

	if err := waitForPort(ctx, rtspPort, 5*time.Second); err != nil {
		cancel()
		return fmt.Errorf("mediamtx ready: %w", err)
	}

	go func() {
		if err := rs.mediamtxCmd.Wait(); err != nil {
			log.Printf("mediamtx exited: %v", err)
		}
		os.RemoveAll(configDir)
	}()

	log.Printf("RTSP: mediamtx listening on :%d", rtspPort)
	return nil
}

func (rs *rtspServer) startWithFrames(frameCh <-chan *image.RGBA, cfg Config) error {
	w := cfg.VideoWidth
	if w == 0 {
		w = DefaultVideoWidth
	}
	h := cfg.VideoHeight
	if h == 0 {
		h = DefaultVideoHeight
	}
	fps := cfg.Framerate
	if fps == 0 {
		fps = DefaultFramerate
	}

	rtspPort := cfg.RtspPort
	if rtspPort == 0 {
		rtspPort = DefaultRtspPort
	}

	ctx, cancel := context.WithCancel(context.Background())

	ffmpegArgs := []string{
		"-f", "rawvideo",
		"-pix_fmt", "rgba",
		"-s", fmt.Sprintf("%dx%d", w, h),
		"-r", fmt.Sprintf("%d", fps),
		"-i", "pipe:0",
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-tune", "zerolatency",
		"-pix_fmt", "yuv420p",
		"-g", fmt.Sprintf("%d", fps),
		"-f", "rtsp",
		"-rtsp_transport", "tcp",
		fmt.Sprintf("rtsp://127.0.0.1:%d/stream", rtspPort),
	}

	rs.ffmpegCmd = exec.CommandContext(ctx, rs.ffmpegPath, ffmpegArgs...)

	ffmpegStdin, err := rs.ffmpegCmd.StdinPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("ffmpeg stdin: %w", err)
	}
	ffmpegStderr, err := rs.ffmpegCmd.StderrPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("ffmpeg stderr: %w", err)
	}

	if err := rs.ffmpegCmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("ffmpeg start: %w", err)
	}

	// Store cancel so close() can kill the pipeline
	rs.mu.Lock()
	rs.cancel = cancel
	rs.mu.Unlock()

	go func() {
		defer ffmpegStdin.Close()
		for frame := range frameCh {
			if _, err := ffmpegStdin.Write(frame.Pix); err != nil {
				return
			}
		}
	}()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ffmpegStderr.Read(buf)
			if err != nil {
				return
			}
			if n > 0 {
				log.Printf("ffmpeg: %s", buf[:n])
			}
		}
	}()

	go func() {
		if err := rs.ffmpegCmd.Wait(); err != nil {
			log.Printf("ffmpeg exited: %v", err)
		}
	}()

	rs.mu.Lock()
	rs.started = true
	rs.mu.Unlock()

	log.Printf("RTSP: streaming to :%d/stream (%dx%d @ %dfps)", rtspPort, w, h, fps)
	return nil
}

func (rs *rtspServer) close() {
	rs.mu.Lock()
	defer rs.mu.Unlock()

	if rs.cancel != nil {
		rs.cancel()
		rs.cancel = nil
	}

	if rs.ffmpegCmd != nil && rs.ffmpegCmd.Process != nil {
		rs.ffmpegCmd.Process.Signal(os.Interrupt)
	}
	if rs.mediamtxCmd != nil && rs.mediamtxCmd.Process != nil {
		rs.mediamtxCmd.Process.Signal(os.Interrupt)
	}

	// Wait briefly for graceful shutdown
	time.Sleep(500 * time.Millisecond)

	if rs.ffmpegCmd != nil && rs.ffmpegCmd.Process != nil {
		rs.ffmpegCmd.Process.Kill()
		rs.ffmpegCmd.Wait()
	}
	if rs.mediamtxCmd != nil && rs.mediamtxCmd.Process != nil {
		rs.mediamtxCmd.Process.Kill()
		rs.mediamtxCmd.Wait()
	}

	rs.started = false
}

func waitForPort(ctx context.Context, port int, timeout time.Duration) error {
	deadline := time.After(timeout)
	addr := fmt.Sprintf("localhost:%d", port)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-deadline:
			return fmt.Errorf("port %d not ready after %v", port, timeout)
		default:
		}
		c, err := (&net.Dialer{Timeout: 200 * time.Millisecond}).DialContext(ctx, "tcp", addr)
		if err == nil {
			c.Close()
			return nil
		}
		time.Sleep(200 * time.Millisecond)
	}
}
