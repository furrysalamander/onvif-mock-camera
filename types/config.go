package types

import (
	"os"
	"strconv"
)

type Config struct {
	DeviceName      string
	DeviceUUID      string
	Manufacturer    string
	Model           string
	FirmwareVersion string

	HostIP    string
	OnvifPort int
	RtspPort  int

	VideoWidth  int
	VideoHeight int
	Framerate   int

	Source VideoSource

	FFmpegPath   string
	MediamtxPath string

	Verbose bool
}

const (
	DefaultDeviceName      = "Mock Camera"
	DefaultManufacturer    = "MockCameraFactory"
	DefaultModel           = "MC-X1"
	DefaultFirmwareVersion = "1.0.0"
	DefaultHostIP          = "127.0.0.1"
	DefaultOnvifPort       = 8080
	DefaultRtspPort        = 8554
	DefaultVideoWidth      = 320
	DefaultVideoHeight     = 200
	DefaultFramerate       = 35
)

func ConfigFromEnv() Config {
	return Config{
		DeviceName:      envStr("ONVIF_DEVICE_NAME", DefaultDeviceName),
		DeviceUUID:      envStr("ONVIF_DEVICE_UUID", ""),
		Manufacturer:    envStr("ONVIF_MANUFACTURER", DefaultManufacturer),
		Model:           envStr("ONVIF_MODEL", DefaultModel),
		FirmwareVersion: envStr("ONVIF_FIRMWARE_VERSION", DefaultFirmwareVersion),
		HostIP:          envStr("ONVIF_HOST_IP", DefaultHostIP),
		OnvifPort:       envInt("ONVIF_PORT", DefaultOnvifPort),
		RtspPort:        envInt("RTSP_PORT", DefaultRtspPort),
		VideoWidth:      envInt("ONVIF_VIDEO_WIDTH", DefaultVideoWidth),
		VideoHeight:     envInt("ONVIF_VIDEO_HEIGHT", DefaultVideoHeight),
		Framerate:       envInt("ONVIF_FRAMERATE", DefaultFramerate),
		FFmpegPath:      envStr("FFMPEG_PATH", "ffmpeg"),
		MediamtxPath:    envStr("MEDIAMTX_PATH", "mediamtx"),
	}
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return def
}
