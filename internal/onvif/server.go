package onvif

import (
	"github.com/furrysalamander/onvif-mock-camera/types"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

func StartServer(addr string, cfg types.Config, source types.VideoSource) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/onvif/device_service", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		resp, err := handleDeviceService(bodyBytes, cfg)
		if err != nil {
			log.Printf("device service error: %v", err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	})

	mux.HandleFunc("/onvif/media_service", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		resp, err := handleMediaService(bodyBytes, cfg)
		if err != nil {
			log.Printf("media service error: %v", err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	})

	mux.HandleFunc("/onvif/ptz_service", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		resp, err := handlePTZService(bodyBytes, source)
		if err != nil {
			log.Printf("PTZ service error: %v", err)
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(resp)
	})

	srv := &http.Server{Addr: addr, Handler: mux}
	go func() {
		log.Printf("ONVIF server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("ONVIF server error: %v", err)
		}
	}()
	return srv
}

func handleDeviceService(bodyBytes []byte, cfg types.Config) ([]byte, error) {
	nsDecls, bodyXML, err := parseEnvelope(bodyBytes)
	if err != nil {
		return nil, err
	}
	inner, err := handleDevice(bodyXML, cfg)
	if err != nil {
		return nil, err
	}
	return buildResponse(nsDecls, inner)
}

func handleMediaService(bodyBytes []byte, cfg types.Config) ([]byte, error) {
	nsDecls, bodyXML, err := parseEnvelope(bodyBytes)
	if err != nil {
		return nil, err
	}
	inner, err := handleMedia(bodyXML, cfg)
	if err != nil {
		return nil, err
	}
	return buildResponse(nsDecls, inner)
}

func handlePTZService(bodyBytes []byte, source types.VideoSource) ([]byte, error) {
	nsDecls, bodyXML, err := parseEnvelope(bodyBytes)
	if err != nil {
		return nil, err
	}
	inner, err := handlePTZ(nsDecls, bodyXML, source)
	if err != nil {
		return nil, err
	}
	return buildResponse(nsDecls, inner)
}

func parseEnvelope(data []byte) (nsDecls string, bodyXML []byte, err error) {
	var env struct {
		Body struct {
			InnerXML []byte `xml:",innerxml"`
		} `xml:"http://www.w3.org/2003/05/soap-envelope Body"`
	}
	if err := xml.Unmarshal(data, &env); err != nil {
		return "", nil, fmt.Errorf("soap parse: %w", err)
	}
	bodyXML = env.Body.InnerXML

	dec := xml.NewDecoder(bytes.NewReader(data))
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == "Envelope" {
			for _, attr := range se.Attr {
				if attr.Name.Local == "" || attr.Name.Local == "xmlns" {
					continue
				}
				if attr.Name.Space == "xmlns" {
					nsDecls += fmt.Sprintf(" xmlns:%s=\"%s\"", attr.Name.Local, attr.Value)
				}
			}
			break
		}
	}
	return nsDecls, bodyXML, nil
}

func unmarshalWithNS(nsDecls string, data []byte, v any) error {
	decls := nsDecls
	if decls == "" {
		decls = ` xmlns:tt="http://www.onvif.org/ver10/schema" xmlns:tds="http://www.onvif.org/ver10/device/wsdl" xmlns:trt="http://www.onvif.org/ver10/media/wsdl" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl"`
	}
	wrapped := []byte(fmt.Sprintf("<root%s>%s</root>", decls, string(data)))

	dec := xml.NewDecoder(bytes.NewReader(wrapped))
	if _, err := dec.Token(); err != nil {
		return fmt.Errorf("skip root: %w", err)
	}
	return dec.Decode(v)
}

func buildResponse(nsDecls string, inner any) ([]byte, error) {
	innerXML, err := xml.Marshal(inner)
	if err != nil {
		return nil, fmt.Errorf("marshal inner: %w", err)
	}

	// soap namespace is hardcoded in the template — strip it from nsDecls to avoid duplication
	cleanNS := strings.ReplaceAll(nsDecls, ` xmlns:soap="http://www.w3.org/2003/05/soap-envelope"`, "")

	if cleanNS == "" {
		cleanNS = ` xmlns:tt="http://www.onvif.org/ver10/schema" xmlns:tds="http://www.onvif.org/ver10/device/wsdl" xmlns:trt="http://www.onvif.org/ver10/media/wsdl" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl"`
	}

	xmlStr := fmt.Sprintf(
		`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope"%s>
    <soap:Header></soap:Header>
    <soap:Body>%s</soap:Body>
</soap:Envelope>`,
		cleanNS,
		string(innerXML),
	)

	return []byte(xmlStr), nil
}
