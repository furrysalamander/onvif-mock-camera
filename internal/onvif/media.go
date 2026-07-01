package onvif

import (
	"github.com/furrysalamander/onvif-mock-camera/types"
	"encoding/xml"
	"fmt"
)

type GetProfilesResponse struct {
	XMLName  xml.Name  `xml:"http://www.onvif.org/ver10/media/wsdl GetProfilesResponse"`
	Profiles []Profile `xml:"http://www.onvif.org/ver10/media/wsdl Profiles"`
}

type GetStreamUriResponse struct {
	XMLName  xml.Name `xml:"http://www.onvif.org/ver10/media/wsdl GetStreamUriResponse"`
	MediaUri MediaUri `xml:"http://www.onvif.org/ver10/media/wsdl MediaUri"`
}

func handleMedia(bodyXML []byte, cfg types.Config) (any, error) {
	op, err := getOperationName(bodyXML)
	if err != nil {
		return nil, err
	}

	w := cfg.VideoWidth
	if w == 0 {
		w = types.DefaultVideoWidth
	}
	h := cfg.VideoHeight
	if h == 0 {
		h = types.DefaultVideoHeight
	}
	hostIP := cfg.HostIP
	if hostIP == "" {
		hostIP = types.DefaultHostIP
	}
	rtspPort := fmt.Sprintf("%d", cfg.RtspPort)
	if cfg.RtspPort == 0 {
		rtspPort = fmt.Sprintf("%d", types.DefaultRtspPort)
	}

	switch op {
	case "GetProfiles":
		return GetProfilesResponse{
			Profiles: []Profile{
				{
					Token: "Profile1",
					Fixed: true,
					Name:  "MainStream",
					VideoEncoderConfiguration: &VideoEncoderConfiguration{
						Token:    "VideoEncoder1",
						Name:     "H264",
						Encoding: "H264",
						Resolution: VideoResolution{
							Width:  w,
							Height: h,
						},
					},
					PTZConfiguration: &PTZConfiguration{
						Token:     "PTZCfg1",
						Name:      "PTZ",
						NodeToken: "PTZNode1",
					},
				},
			},
		}, nil

	case "GetStreamUri":
		return GetStreamUriResponse{
			MediaUri: MediaUri{
				URI:                 fmt.Sprintf("rtsp://%s:%s/stream", hostIP, rtspPort),
				InvalidAfterConnect: false,
				InvalidAfterReboot:  false,
				Timeout:             "PT30S",
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown media operation: %s", op)
	}
}
