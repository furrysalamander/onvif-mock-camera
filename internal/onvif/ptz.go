package onvif

import (
	"github.com/furrysalamander/onvif-mock-camera/types"
	"encoding/xml"
	"fmt"
)

type ContinuousMoveRequest struct {
	XMLName      xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl ContinuousMove"`
	ProfileToken string   `xml:"http://www.onvif.org/ver20/ptz/wsdl ProfileToken"`
	Velocity     PTZSpeed `xml:"http://www.onvif.org/ver20/ptz/wsdl Velocity"`
	Timeout      string   `xml:"http://www.onvif.org/ver20/ptz/wsdl Timeout,omitempty"`
}

type AbsoluteMoveRequest struct {
	XMLName      xml.Name  `xml:"http://www.onvif.org/ver20/ptz/wsdl AbsoluteMove"`
	ProfileToken string    `xml:"http://www.onvif.org/ver20/ptz/wsdl ProfileToken"`
	Position     PTZVector `xml:"http://www.onvif.org/ver20/ptz/wsdl Position"`
	Speed        *PTZSpeed `xml:"http://www.onvif.org/ver20/ptz/wsdl Speed,omitempty"`
}

type RelativeMoveRequest struct {
	XMLName      xml.Name  `xml:"http://www.onvif.org/ver20/ptz/wsdl RelativeMove"`
	ProfileToken string    `xml:"http://www.onvif.org/ver20/ptz/wsdl ProfileToken"`
	Translation  PTZVector `xml:"http://www.onvif.org/ver20/ptz/wsdl Translation"`
	Speed        *PTZSpeed `xml:"http://www.onvif.org/ver20/ptz/wsdl Speed,omitempty"`
}

type StopRequest struct {
	XMLName      xml.Name `xml:"http://www.onvif.org/ver20/ptz/wsdl Stop"`
	ProfileToken string   `xml:"http://www.onvif.org/ver20/ptz/wsdl ProfileToken"`
	PanTilt      bool     `xml:"http://www.onvif.org/ver20/ptz/wsdl PanTilt"`
	Zoom         bool     `xml:"http://www.onvif.org/ver20/ptz/wsdl Zoom"`
}

type GetConfigurationsResponse struct {
	XMLName          xml.Name           `xml:"http://www.onvif.org/ver20/ptz/wsdl GetConfigurationsResponse"`
	PTZConfiguration []PTZConfiguration `xml:"http://www.onvif.org/ver20/ptz/wsdl PTZConfiguration"`
}

type ptzEmptyResponse struct {
	XMLName xml.Name
}

func handlePTZ(nsDecls string, bodyXML []byte, source types.VideoSource) (any, error) {
	op, err := getOperationName(bodyXML)
	if err != nil {
		return nil, err
	}

	switch op {
	case "ContinuousMove":
		var req ContinuousMoveRequest
		if err := unmarshalWithNS(nsDecls, bodyXML, &req); err != nil {
			return nil, fmt.Errorf("ContinuousMove parse: %w", err)
		}
		pan := 0.0
		tilt := 0.0
		zoom := 0.0
		if req.Velocity.PanTilt != nil {
			pan = req.Velocity.PanTilt.X
			tilt = req.Velocity.PanTilt.Y
		}
		if req.Velocity.Zoom != nil {
			zoom = req.Velocity.Zoom.X
		}
		if err := source.ContinuousMove(pan, tilt, zoom); err != nil {
			return nil, fmt.Errorf("ContinuousMove: %w", err)
		}
		return ptzEmptyResponse{XMLName: xml.Name{Local: "ContinuousMoveResponse", Space: PTZNS}}, nil

	case "AbsoluteMove":
		var req AbsoluteMoveRequest
		if err := unmarshalWithNS(nsDecls, bodyXML, &req); err != nil {
			return nil, fmt.Errorf("AbsoluteMove parse: %w", err)
		}
		pan := 0.0
		tilt := 0.0
		zoom := 0.0
		if req.Position.PanTilt != nil {
			pan = req.Position.PanTilt.X
			tilt = req.Position.PanTilt.Y
		}
		if req.Position.Zoom != nil {
			zoom = req.Position.Zoom.X
		}
		if err := source.AbsoluteMove(pan, tilt, zoom); err != nil {
			return nil, fmt.Errorf("AbsoluteMove: %w", err)
		}
		return ptzEmptyResponse{XMLName: xml.Name{Local: "AbsoluteMoveResponse", Space: PTZNS}}, nil

	case "RelativeMove":
		var req RelativeMoveRequest
		if err := unmarshalWithNS(nsDecls, bodyXML, &req); err != nil {
			return nil, fmt.Errorf("RelativeMove parse: %w", err)
		}
		pan := 0.0
		tilt := 0.0
		zoom := 0.0
		if req.Translation.PanTilt != nil {
			pan = req.Translation.PanTilt.X
			tilt = req.Translation.PanTilt.Y
		}
		if req.Translation.Zoom != nil {
			zoom = req.Translation.Zoom.X
		}
		if err := source.RelativeMove(pan, tilt, zoom); err != nil {
			return nil, fmt.Errorf("RelativeMove: %w", err)
		}
		return ptzEmptyResponse{XMLName: xml.Name{Local: "RelativeMoveResponse", Space: PTZNS}}, nil

	case "Stop":
		var req StopRequest
		if err := unmarshalWithNS(nsDecls, bodyXML, &req); err != nil {
			return nil, fmt.Errorf("Stop parse: %w", err)
		}
		if err := source.StopPTZ(req.PanTilt, req.Zoom); err != nil {
			return nil, fmt.Errorf("Stop: %w", err)
		}
		return ptzEmptyResponse{XMLName: xml.Name{Local: "StopResponse", Space: PTZNS}}, nil

	case "GetConfigurations":
		return GetConfigurationsResponse{
			PTZConfiguration: []PTZConfiguration{
				{
					Token:     "PTZCfg1",
					Name:      "PTZ",
					NodeToken: "PTZNode1",
				},
			},
		}, nil

	default:
		return nil, fmt.Errorf("unknown PTZ operation: %s", op)
	}
}
