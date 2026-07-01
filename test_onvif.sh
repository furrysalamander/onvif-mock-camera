#!/bin/bash
# Test all ONVIF endpoints against a running mock camera.
# Usage: ./test_onvif.sh [host] [onvif_port] [rtsp_port]
# Defaults: host=127.0.0.1, onvif_port=8080, rtsp_port=8554

HOST="${1:-127.0.0.1}"
ONVIF_PORT="${2:-8080}"
RTSP_PORT="${3:-8554}"
BASE="http://${HOST}:${ONVIF_PORT}"

die() { echo "FAIL: $*"; exit 1; }

check() {
    local label="$1" url="$2" body="$3" want="$4"
    local resp
    resp=$(curl -s -X POST "$url" -H 'Content-Type: application/soap+xml' -d "$body")
    if echo "$resp" | grep -q "$want"; then
        echo "  OK   $label"
    else
        echo "  FAIL $label"
        echo "       wanted: $want"
        echo "       got:    $resp"
    fi
}

echo "ONVIF mock camera endpoint tests"
echo "  Host: ${HOST}"
echo "  ONVIF port: ${ONVIF_PORT}"
echo "  RTSP port:  ${RTSP_PORT}"
echo ""

# --- Device service ---
echo "=== Device service ==="

check "GetDeviceInformation" \
    "${BASE}/onvif/device_service" \
    '<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tds="http://www.onvif.org/ver10/device/wsdl"><soap:Body><tds:GetDeviceInformation/></soap:Body></soap:Envelope>' \
    'GetDeviceInformationResponse'

check "GetSystemDateAndTime" \
    "${BASE}/onvif/device_service" \
    '<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tds="http://www.onvif.org/ver10/device/wsdl"><soap:Body><tds:GetSystemDateAndTime/></soap:Body></soap:Envelope>' \
    'GetSystemDateAndTimeResponse'

check "GetCapabilities" \
    "${BASE}/onvif/device_service" \
    '<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tds="http://www.onvif.org/ver10/device/wsdl"><soap:Body><tds:GetCapabilities/></soap:Body></soap:Envelope>' \
    'GetCapabilitiesResponse'

# --- Media service ---
echo "=== Media service ==="

check "GetProfiles" \
    "${BASE}/onvif/media_service" \
    '<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:trt="http://www.onvif.org/ver10/media/wsdl"><soap:Body><trt:GetProfiles/></soap:Body></soap:Envelope>' \
    'GetProfilesResponse'

check "GetStreamUri" \
    "${BASE}/onvif/media_service" \
    '<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:trt="http://www.onvif.org/ver10/media/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema"><soap:Body><trt:GetStreamUri><trt:StreamSetup><tt:Stream>RTP-Unicast</tt:Stream><tt:Transport><tt:Protocol>RTSP</tt:Protocol></tt:Transport></trt:StreamSetup><trt:ProfileToken>Profile1</trt:ProfileToken></trt:GetStreamUri></soap:Body></soap:Envelope>' \
    "rtsp://${HOST}:${RTSP_PORT}/stream"

# --- PTZ service ---
echo "=== PTZ service ==="

check "GetConfigurations" \
    "${BASE}/onvif/ptz_service" \
    '<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl"><soap:Body><tptz:GetConfigurations/></soap:Body></soap:Envelope>' \
    'GetConfigurationsResponse'

# ContinuousMove: pan right at 50% speed
check "ContinuousMove (pan right)" \
    "${BASE}/onvif/ptz_service" \
    '<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema"><soap:Body><tptz:ContinuousMove><tptz:ProfileToken>Profile1</tptz:ProfileToken><tptz:Velocity><tt:PanTilt x="0.5" y="0.0"/><tt:Zoom x="0.0"/></tptz:Velocity></tptz:ContinuousMove></soap:Body></soap:Envelope>' \
    'ContinuousMoveResponse'

# ContinuousMove: move forward
check "ContinuousMove (forward)" \
    "${BASE}/onvif/ptz_service" \
    '<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema"><soap:Body><tptz:ContinuousMove><tptz:ProfileToken>Profile1</tptz:ProfileToken><tptz:Velocity><tt:PanTilt x="0.0" y="1.0"/><tt:Zoom x="0.0"/></tptz:Velocity></tptz:ContinuousMove></soap:Body></soap:Envelope>' \
    'ContinuousMoveResponse'

check "Stop" \
    "${BASE}/onvif/ptz_service" \
    '<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl"><soap:Body><tptz:Stop><tptz:ProfileToken>Profile1</tptz:ProfileToken><tptz:PanTilt>true</tptz:PanTilt><tptz:Zoom>true</tptz:Zoom></tptz:Stop></soap:Body></soap:Envelope>' \
    'StopResponse'

# RelativeMove: fire weapon (positive zoom)
check "RelativeMove (fire)" \
    "${BASE}/onvif/ptz_service" \
    '<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema"><soap:Body><tptz:RelativeMove><tptz:ProfileToken>Profile1</tptz:ProfileToken><tptz:Translation><tt:PanTilt x="0.0" y="0.0"/><tt:Zoom x="1.0"/></tptz:Translation></tptz:RelativeMove></soap:Body></soap:Envelope>' \
    'RelativeMoveResponse'

# RelativeMove: use/open (negative zoom)
check "RelativeMove (use)" \
    "${BASE}/onvif/ptz_service" \
    '<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema"><soap:Body><tptz:RelativeMove><tptz:ProfileToken>Profile1</tptz:ProfileToken><tptz:Translation><tt:PanTilt x="0.0" y="0.0"/><tt:Zoom x="-1.0"/></tptz:Translation></tptz:RelativeMove></soap:Body></soap:Envelope>' \
    'RelativeMoveResponse'

# AbsoluteMove (no-op but should return OK)
check "AbsoluteMove" \
    "${BASE}/onvif/ptz_service" \
    '<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope" xmlns:tptz="http://www.onvif.org/ver20/ptz/wsdl" xmlns:tt="http://www.onvif.org/ver10/schema"><soap:Body><tptz:AbsoluteMove><tptz:ProfileToken>Profile1</tptz:ProfileToken><tptz:Position><tt:PanTilt x="0.5" y="0.5"/><tt:Zoom x="0.0"/></tptz:Position></tptz:AbsoluteMove></soap:Body></soap:Envelope>' \
    'AbsoluteMoveResponse'

echo ""
echo "Done."
