package onvifmock

import (
	"image"
	"strings"
	"testing"
)

type testSource struct {
	pan, tilt, zoom       float64
	stopPT, stopZoom      bool
	relMoves              [][3]float64
	absMoves              [][3]float64
	frameSize             [2]int
}

func (ts *testSource) Start() (<-chan *image.RGBA, error) { return nil, nil }
func (ts *testSource) Stop() error                        { return nil }
func (ts *testSource) FrameSize() (int, int)              { return ts.frameSize[0], ts.frameSize[1] }
func (ts *testSource) ContinuousMove(pan, tilt, zoom float64) error {
	ts.pan, ts.tilt, ts.zoom = pan, tilt, zoom
	return nil
}
func (ts *testSource) RelativeMove(pan, tilt, zoom float64) error {
	ts.relMoves = append(ts.relMoves, [3]float64{pan, tilt, zoom})
	return nil
}
func (ts *testSource) AbsoluteMove(pan, tilt, zoom float64) error {
	ts.absMoves = append(ts.absMoves, [3]float64{pan, tilt, zoom})
	return nil
}
func (ts *testSource) StopPTZ(panTilt, zoom bool) error {
	ts.stopPT, ts.stopZoom = panTilt, zoom
	return nil
}

func TestDeviceService(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCont string
	}{
		{
			"GetDeviceInformation",
			`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tds="http://www.onvif.org/ver10/device/wsdl"><soap:Body><tds:GetDeviceInformation/></soap:Body></soap:Envelope>`,
			"GetDeviceInformationResponse",
		},
		{
			"GetCapabilities",
			`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tds="http://www.onvif.org/ver10/device/wsdl"><soap:Body><tds:GetCapabilities/></soap:Body></soap:Envelope>`,
			"GetCapabilitiesResponse",
		},
		{
			"GetSystemDateAndTime",
			`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tds="http://www.onvif.org/ver10/device/wsdl"><soap:Body><tds:GetSystemDateAndTime/></soap:Body></soap:Envelope>`,
			"GetSystemDateAndTimeResponse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handleDeviceService([]byte(tt.body), Config{})
			if err != nil {
				t.Fatalf("handleDeviceService: %v", err)
			}
			if !strings.Contains(string(resp), tt.wantCont) {
				t.Errorf("response missing %q: %s", tt.wantCont, string(resp))
			}
		})
	}
}

func TestMediaService(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		wantCont string
	}{
		{
			"GetProfiles",
			`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:trt="http://www.onvif.org/ver10/media/wsdl"><soap:Body><trt:GetProfiles/></soap:Body></soap:Envelope>`,
			"GetProfilesResponse",
		},
		{
			"GetStreamUri",
			`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:trt="http://www.onvif.org/ver10/media/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema"><soap:Body><trt:GetStreamUri><trt:StreamSetup><tt:Stream>RTP-Unicast</tt:Stream><tt:Transport><tt:Protocol>RTSP</tt:Protocol></tt:Transport></trt:StreamSetup><trt:ProfileToken>Profile1</trt:ProfileToken></trt:GetStreamUri></soap:Body></soap:Envelope>`,
			"GetStreamUriResponse",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handleMediaService([]byte(tt.body), Config{})
			if err != nil {
				t.Fatalf("handleMediaService: %v", err)
			}
			if !strings.Contains(string(resp), tt.wantCont) {
				t.Errorf("response missing %q: %s", tt.wantCont, string(resp))
			}
		})
	}
}

func TestPTZService(t *testing.T) {
	ts := &testSource{}

	tests := []struct {
		name     string
		body     string
		check    func(t *testing.T)
	}{
		{
			"ContinuousMove",
			`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema"><soap:Body><tptz:ContinuousMove><tptz:ProfileToken>Profile1</tptz:ProfileToken><tptz:Velocity><tt:PanTilt x="0.5" y="-0.3"/><tt:Zoom x="0.8"/></tptz:Velocity></tptz:ContinuousMove></soap:Body></soap:Envelope>`,
			func(t *testing.T) {
				if ts.pan != 0.5 || ts.tilt != -0.3 || ts.zoom != 0.8 {
					t.Errorf("ContinuousMove: got pan=%v tilt=%v zoom=%v", ts.pan, ts.tilt, ts.zoom)
				}
			},
		},
		{
			"RelativeMove",
			`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema"><soap:Body><tptz:RelativeMove><tptz:ProfileToken>Profile1</tptz:ProfileToken><tptz:Translation><tt:PanTilt x="0.2" y="0.1"/><tt:Zoom x="-0.5"/></tptz:Translation></tptz:RelativeMove></soap:Body></soap:Envelope>`,
			func(t *testing.T) {
				if len(ts.relMoves) == 0 {
					t.Fatal("no RelativeMove recorded")
				}
				last := ts.relMoves[len(ts.relMoves)-1]
				if last[0] != 0.2 || last[1] != 0.1 || last[2] != -0.5 {
					t.Errorf("RelativeMove: got [%v,%v,%v]", last[0], last[1], last[2])
				}
			},
		},
		{
			"Stop",
			`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl"><soap:Body><tptz:Stop><tptz:ProfileToken>Profile1</tptz:ProfileToken><tptz:PanTilt>true</tptz:PanTilt><tptz:Zoom>false</tptz:Zoom></tptz:Stop></soap:Body></soap:Envelope>`,
			func(t *testing.T) {
				if ts.stopPT != true || ts.stopZoom != false {
					t.Errorf("Stop: got panTilt=%v zoom=%v", ts.stopPT, ts.stopZoom)
				}
			},
		},
		{
			"GetConfigurations",
			`<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl"><soap:Body><tptz:GetConfigurations/></soap:Body></soap:Envelope>`,
			func(t *testing.T) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := handlePTZService([]byte(tt.body), ts)
			if err != nil {
				t.Fatalf("handlePTZService: %v", err)
			}
			_ = resp
			tt.check(t)
		})
	}
}

func TestParseEnvelope(t *testing.T) {
	body := `<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema"><soap:Body><tptz:ContinuousMove><tptz:ProfileToken>P1</tptz:ProfileToken></tptz:ContinuousMove></soap:Body></soap:Envelope>`
	nsDecls, bodyXML, err := parseEnvelope([]byte(body))
	if err != nil {
		t.Fatalf("parseEnvelope: %v", err)
	}
	if !strings.Contains(nsDecls, "tptz") {
		t.Errorf("expected tptz in nsDecls, got: %s", nsDecls)
	}
	if !strings.Contains(nsDecls, "tt") {
		t.Errorf("expected tt in nsDecls, got: %s", nsDecls)
	}
	if !strings.Contains(nsDecls, "soap") {
		t.Errorf("expected soap in nsDecls, got: %s", nsDecls)
	}
	if !strings.Contains(string(bodyXML), "ContinuousMove") {
		t.Errorf("expected ContinuousMove in bodyXML, got: %s", string(bodyXML))
	}
}

func TestUnmarshalWithNS(t *testing.T) {
	bodyXML := []byte(`<tptz:ContinuousMove><tptz:ProfileToken>Profile1</tptz:ProfileToken><tptz:Velocity><tt:PanTilt x="0.5" y="-0.3"/><tt:Zoom x="0.8"/></tptz:Velocity></tptz:ContinuousMove>`)
	nsDecls := ` xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema"`

	var req ContinuousMoveRequest
	if err := unmarshalWithNS(nsDecls, bodyXML, &req); err != nil {
		t.Fatalf("unmarshalWithNS: %v", err)
	}
	if req.ProfileToken != "Profile1" {
		t.Errorf("token = %s", req.ProfileToken)
	}
	if req.Velocity.PanTilt == nil {
		t.Fatal("PanTilt is nil")
	}
	if req.Velocity.PanTilt.X != 0.5 || req.Velocity.PanTilt.Y != -0.3 {
		t.Errorf("PanTilt = {%v, %v}", req.Velocity.PanTilt.X, req.Velocity.PanTilt.Y)
	}
	if req.Velocity.Zoom == nil {
		t.Fatal("Zoom is nil")
	}
	if req.Velocity.Zoom.X != 0.8 {
		t.Errorf("Zoom = %v", req.Velocity.Zoom.X)
	}
}

func TestBuildResponse(t *testing.T) {
	resp := GetConfigurationsResponse{
		PTZConfiguration: []PTZConfiguration{
			{Token: "t1", Name: "n1", NodeToken: "nt1"},
		},
	}
	xml, err := buildResponse(` xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema"`, resp)
	if err != nil {
		t.Fatalf("buildResponse: %v", err)
	}
	if !strings.Contains(string(xml), "GetConfigurationsResponse") {
		t.Errorf("missing response element: %s", string(xml))
	}
	if !strings.Contains(string(xml), "tptz") {
		t.Errorf("missing tptz namespace: %s", string(xml))
	}
}
