package onvifmock

import (
	"fmt"
	"log"
	"net"
	"strings"

	"golang.org/x/net/ipv4"
)

func runDiscovery(cfg Config) {
	addr := net.UDPAddr{
		IP:   net.IPv4(239, 255, 255, 250),
		Port: 3702,
	}

	conn, err := net.ListenMulticastUDP("udp4", nil, &addr)
	if err != nil {
		log.Printf("discovery: listen failed: %v", err)
		return
	}
	defer conn.Close()

	pc := ipv4.NewPacketConn(conn)
	if err := pc.SetControlMessage(ipv4.FlagDst, true); err != nil {
		log.Printf("discovery: set control message: %v", err)
	}

	log.Printf("WS-Discovery listening on %s", addr.String())

	buf := make([]byte, 4096)
	for {
		n, _, src, err := pc.ReadFrom(buf)
		if err != nil {
			log.Printf("discovery: read error: %v", err)
			continue
		}

		msg := string(buf[:n])
		if !isProbe(msg) {
			continue
		}

		log.Printf("discovery: received Probe from %s", src)

		resp := buildProbeMatch(msg, cfg)
		if resp == "" {
			continue
		}

		if _, err := conn.WriteTo([]byte(resp), src); err != nil {
			log.Printf("discovery: write response: %v", err)
		} else {
			log.Printf("discovery: sent ProbeMatch to %s", src)
		}
	}
}

func isProbe(msg string) bool {
	return strings.Contains(msg, ":Probe") || strings.Contains(msg, "Probe>")
}

func buildProbeMatch(probeMsg string, cfg Config) string {
	msgID := extractTagContent(probeMsg, "MessageID")
	if msgID == "" {
		return ""
	}

	messageUUID := fmt.Sprintf("urn:uuid:%s", cfg.DeviceUUID)
	if cfg.DeviceUUID == "" {
		messageUUID = "urn:uuid:probe-match-response"
	}
	deviceUUID := cfg.DeviceUUID
	if deviceUUID == "" {
		deviceUUID = "00000000-0000-0000-0000-000000000000"
	}

	hostIP := cfg.HostIP
	if hostIP == "" {
		hostIP = DefaultHostIP
	}
	port := fmt.Sprintf("%d", cfg.OnvifPort)
	if cfg.OnvifPort == 0 {
		port = fmt.Sprintf("%d", DefaultOnvifPort)
	}

	return fmt.Sprintf(
		`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"
               xmlns:wsa="http://schemas.xmlsoap.org/ws/2004/08/addressing"
               xmlns:d="http://schemas.xmlsoap.org/ws/2005/04/discovery"
               xmlns:dn="http://www.onvif.org/ver10/network/wsdl">
    <soap:Header>
        <wsa:To>http://schemas.xmlsoap.org/ws/2004/08/addressing/role/anonymous</wsa:To>
        <wsa:Action>http://schemas.xmlsoap.org/ws/2005/04/discovery/ProbeMatch</wsa:Action>
        <wsa:MessageID>%s</wsa:MessageID>
        <wsa:RelatesTo>%s</wsa:RelatesTo>
    </soap:Header>
    <soap:Body>
        <d:ProbeMatch>
            <d:EndpointReference>
                <wsa:Address>urn:uuid:%s</wsa:Address>
            </d:EndpointReference>
            <d:Types>dn:NetworkVideoTransmitter</d:Types>
            <d:Scopes>onvif://www.onvif.org/type/video_encoder onvif://www.onvif.org/type/audio_encoder onvif://www.onvif.org/hardware/MockCamera onvif://www.onvif.org/name/MockCamera</d:Scopes>
            <d:XAddrs>http://%s:%s/onvif/device_service</d:XAddrs>
            <d:MetadataVersion>1</d:MetadataVersion>
        </d:ProbeMatch>
    </soap:Body>
</soap:Envelope>`,
		messageUUID, msgID, deviceUUID, hostIP, port,
	)
}

func extractTagContent(xml, tagName string) string {
	tag := fmt.Sprintf(":%s", tagName)
	start := strings.Index(xml, tag)
	if start == -1 {
		return ""
	}
	contentStart := strings.Index(xml[start:], ">")
	if contentStart == -1 {
		return ""
	}
	contentStart += start + 1
	endTag := fmt.Sprintf("</%s", tagName)
	end := strings.Index(xml[contentStart:], endTag)
	if end == -1 {
		return ""
	}
	return xml[contentStart : contentStart+end]
}
