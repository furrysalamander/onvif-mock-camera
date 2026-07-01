package onvifmock

import "encoding/xml"

type XMLDateTime struct {
	Time XMLTime `xml:"http://www.onvif.org/ver10/schema Time"`
	Date XMLDate `xml:"http://www.onvif.org/ver10/schema Date"`
}

type XMLTime struct {
	Hour   int `xml:"http://www.onvif.org/ver10/schema Hour"`
	Minute int `xml:"http://www.onvif.org/ver10/schema Minute"`
	Second int `xml:"http://www.onvif.org/ver10/schema Second"`
}

type XMLDate struct {
	Year  int `xml:"http://www.onvif.org/ver10/schema Year"`
	Month int `xml:"http://www.onvif.org/ver10/schema Month"`
	Day   int `xml:"http://www.onvif.org/ver10/schema Day"`
}

type Profile struct {
	XMLName                    xml.Name                    `xml:"http://www.onvif.org/ver10/schema Profile"`
	Token                      string                      `xml:"token,attr"`
	Fixed                      bool                        `xml:"fixed,attr"`
	Name                       string                      `xml:"http://www.onvif.org/ver10/schema Name"`
	VideoEncoderConfiguration  *VideoEncoderConfiguration  `xml:"http://www.onvif.org/ver10/schema VideoEncoderConfiguration"`
	PTZConfiguration           *PTZConfiguration            `xml:"http://www.onvif.org/ver10/schema PTZConfiguration"`
}

type VideoEncoderConfiguration struct {
	Token      string          `xml:"token,attr"`
	Name       string          `xml:"http://www.onvif.org/ver10/schema Name"`
	Encoding   string          `xml:"http://www.onvif.org/ver10/schema Encoding"`
	Resolution VideoResolution `xml:"http://www.onvif.org/ver10/schema Resolution"`
}

type VideoResolution struct {
	Width  int `xml:"http://www.onvif.org/ver10/schema Width"`
	Height int `xml:"http://www.onvif.org/ver10/schema Height"`
}

type PTZConfiguration struct {
	Token     string `xml:"token,attr"`
	Name      string `xml:"http://www.onvif.org/ver10/schema Name"`
	NodeToken string `xml:"http://www.onvif.org/ver10/schema NodeToken"`
}

type StreamSetup struct {
	Stream    string    `xml:"http://www.onvif.org/ver10/schema Stream"`
	Transport Transport `xml:"http://www.onvif.org/ver10/schema Transport"`
}

type Transport struct {
	Protocol string `xml:"http://www.onvif.org/ver10/schema Protocol"`
}

type MediaUri struct {
	URI                    string `xml:"http://www.onvif.org/ver10/schema Uri"`
	InvalidAfterConnect    bool   `xml:"http://www.onvif.org/ver10/schema InvalidAfterConnect"`
	InvalidAfterReboot     bool   `xml:"http://www.onvif.org/ver10/schema InvalidAfterReboot"`
	Timeout                string `xml:"http://www.onvif.org/ver10/schema Timeout"`
}

type PTZSpeed struct {
	PanTilt *Vector2D `xml:"http://www.onvif.org/ver10/schema PanTilt"`
	Zoom    *Vector1D `xml:"http://www.onvif.org/ver10/schema Zoom"`
}

type Vector2D struct {
	X     float64 `xml:"x,attr"`
	Y     float64 `xml:"y,attr"`
	Space string  `xml:"space,attr,omitempty"`
}

type Vector1D struct {
	X     float64 `xml:"x,attr"`
	Space string  `xml:"space,attr,omitempty"`
}

type PTZVector struct {
	PanTilt *Vector2D `xml:"http://www.onvif.org/ver10/schema PanTilt"`
	Zoom    *Vector1D `xml:"http://www.onvif.org/ver10/schema Zoom"`
}
