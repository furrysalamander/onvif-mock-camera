package onvif

import (
	"github.com/furrysalamander/onvif-mock-camera/types"
	"bytes"
	"encoding/xml"
	"fmt"
	"time"
)

type SystemDateAndTime struct {
	XMLName         xml.Name    `xml:"http://www.onvif.org/ver10/schema SystemDateAndTime"`
	DateTimeType    string      `xml:"http://www.onvif.org/ver10/schema DateTimeType"`
	DaylightSavings bool        `xml:"http://www.onvif.org/ver10/schema DaylightSavings"`
	TimeZone        TZ          `xml:"http://www.onvif.org/ver10/schema TimeZone"`
	UTCDateTime     XMLDateTime `xml:"http://www.onvif.org/ver10/schema UTCDateTime"`
}

type TZ struct {
	TZ string `xml:"http://www.onvif.org/ver10/schema TZ"`
}

type GetDeviceInformationResponse struct {
	XMLName         xml.Name `xml:"http://www.onvif.org/ver10/device/wsdl GetDeviceInformationResponse"`
	Manufacturer    string   `xml:"http://www.onvif.org/ver10/device/wsdl Manufacturer"`
	Model           string   `xml:"http://www.onvif.org/ver10/device/wsdl Model"`
	FirmwareVersion string   `xml:"http://www.onvif.org/ver10/device/wsdl FirmwareVersion"`
	SerialNumber    string   `xml:"http://www.onvif.org/ver10/device/wsdl SerialNumber"`
	HardwareId      string   `xml:"http://www.onvif.org/ver10/device/wsdl HardwareId"`
}

type GetSystemDateAndTimeResponse struct {
	XMLName           xml.Name          `xml:"http://www.onvif.org/ver10/device/wsdl GetSystemDateAndTimeResponse"`
	SystemDateAndTime SystemDateAndTime `xml:"http://www.onvif.org/ver10/device/wsdl SystemDateAndTime"`
}

type GetCapabilitiesResponse struct {
	XMLName      xml.Name     `xml:"http://www.onvif.org/ver10/device/wsdl GetCapabilitiesResponse"`
	Capabilities Capabilities `xml:"http://www.onvif.org/ver10/device/wsdl Capabilities"`
}

type Capabilities struct {
	Device *DeviceCapabilities `xml:"http://www.onvif.org/ver10/schema Device"`
	Media  *MediaCapabilities  `xml:"http://www.onvif.org/ver10/schema Media"`
	PTZ    *PTZCapabilities    `xml:"http://www.onvif.org/ver10/schema PTZ"`
}

type DeviceCapabilities struct {
	XAddr string `xml:"http://www.onvif.org/ver10/schema XAddr"`
}

type MediaCapabilities struct {
	XAddr                 string                `xml:"http://www.onvif.org/ver10/schema XAddr"`
	StreamingCapabilities StreamingCapabilities `xml:"http://www.onvif.org/ver10/schema StreamingCapabilities"`
}

type StreamingCapabilities struct {
	RTPMulticast bool `xml:"http://www.onvif.org/ver10/schema RTPMulticast"`
	RTPTCP       bool `xml:"http://www.onvif.org/ver10/schema RTP_TCP"`
	RTPRTSPTCP   bool `xml:"http://www.onvif.org/ver10/schema RTP_RTSP_TCP"`
}

type PTZCapabilities struct {
	XAddr string `xml:"http://www.onvif.org/ver10/schema XAddr"`
}

func handleDevice(bodyXML []byte, cfg types.Config) (any, error) {
	op, err := getOperationName(bodyXML)
	if err != nil {
		return nil, err
	}

	ip := cfg.HostIP
	if ip == "" {
		ip = types.DefaultHostIP
	}
	port := fmt.Sprintf("%d", cfg.OnvifPort)
	if cfg.OnvifPort == 0 {
		port = fmt.Sprintf("%d", types.DefaultOnvifPort)
	}

	switch op {
	case "GetDeviceInformation":
		mfr := cfg.Manufacturer
		if mfr == "" {
			mfr = types.DefaultManufacturer
		}
		model := cfg.Model
		if model == "" {
			model = types.DefaultModel
		}
		fw := cfg.FirmwareVersion
		if fw == "" {
			fw = types.DefaultFirmwareVersion
		}
		uuid := cfg.DeviceUUID
		if uuid == "" {
			uuid = "00000000-0000-0000-0000-000000000000"
		}
		name := cfg.DeviceName
		if name == "" {
			name = types.DefaultDeviceName
		}
		return GetDeviceInformationResponse{
			Manufacturer:    mfr,
			Model:           model,
			FirmwareVersion: fw,
			SerialNumber:    uuid,
			HardwareId:      name,
		}, nil

	case "GetSystemDateAndTime":
		now := time.Now().UTC()
		return GetSystemDateAndTimeResponse{
			SystemDateAndTime: SystemDateAndTime{
				DateTimeType:    "NTP",
				DaylightSavings: false,
				TimeZone:        TZ{TZ: "UTC"},
				UTCDateTime: XMLDateTime{
					Time: XMLTime{
						Hour:   now.Hour(),
						Minute: now.Minute(),
						Second: now.Second(),
					},
					Date: XMLDate{
						Year:  now.Year(),
						Month: int(now.Month()),
						Day:   now.Day(),
					},
				},
			},
		}, nil

	case "GetCapabilities":
		return GetCapabilitiesResponse{
			Capabilities: Capabilities{
				Device: &DeviceCapabilities{
					XAddr: fmt.Sprintf("http://%s:%s/onvif/device_service", ip, port),
				},
				Media: &MediaCapabilities{
					XAddr: fmt.Sprintf("http://%s:%s/onvif/media_service", ip, port),
					StreamingCapabilities: StreamingCapabilities{
						RTPMulticast: false,
						RTPTCP:       true,
						RTPRTSPTCP:   true,
					},
				},
				PTZ: &PTZCapabilities{
					XAddr: fmt.Sprintf("http://%s:%s/onvif/ptz_service", ip, port),
				},
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown device operation: %s", op)
	}
}

func getOperationName(bodyXML []byte) (string, error) {
	dec := xml.NewDecoder(bytes.NewReader(bodyXML))
	for {
		tok, err := dec.Token()
		if err != nil {
			return "", fmt.Errorf("failed to parse body: %w", err)
		}
		if se, ok := tok.(xml.StartElement); ok {
			return se.Name.Local, nil
		}
	}
}
