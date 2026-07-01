package onvif

import "encoding/xml"

const (
	SoapNS     = "http://www.w3.org/2003/05/soap-envelope"
	WsaNS      = "http://schemas.xmlsoap.org/ws/2004/08/addressing"
	WsdNS      = "http://schemas.xmlsoap.org/ws/2005/04/discovery"
	OnvifNetNS = "http://www.onvif.org/ver10/network/wsdl"
	SchemaNS   = "http://www.onvif.org/ver10/schema"
	DeviceNS   = "http://www.onvif.org/ver10/device/wsdl"
	MediaNS    = "http://www.onvif.org/ver10/media/wsdl"
	PTZNS      = "http://www.onvif.org/ver20/ptz/wsdl"
)

type SoapEnvelope struct {
	XMLName xml.Name   `xml:"http://www.w3.org/2003/05/soap-envelope Envelope"`
	Header  SoapHeader `xml:"http://www.w3.org/2003/05/soap-envelope Header"`
	Body    SoapBody   `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
}

type SoapHeader struct {
}

type SoapBody struct {
	InnerXML []byte `xml:",innerxml"`
}

type SoapFault struct {
	XMLName xml.Name       `xml:"http://www.w3.org/2003/05/soap-envelope Fault"`
	Code    SoapFaultCode  `xml:"Code"`
	Reason  SoapFaultReason `xml:"Reason"`
}

type SoapFaultCode struct {
	Value string `xml:"Value"`
}

type SoapFaultReason struct {
	Text SoapFaultText `xml:"Text"`
}

type SoapFaultText struct {
	Lang  string `xml:"http://www.w3.org/XML/1998/namespace lang,attr"`
	Value string `xml:",chardata"`
}
