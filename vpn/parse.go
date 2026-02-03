package vpn

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

var ErrUnsupportedProtocol = errors.New("unsupported protocol")

type vmessPayload struct {
	Version     string `json:"v"`
	Name        string `json:"ps"`
	Address     string `json:"add"`
	Port        string `json:"port"`
	ID          string `json:"id"`
	AlterID     string `json:"aid"`
	Network     string `json:"net"`
	Type        string `json:"type"`
	Host        string `json:"host"`
	Path        string `json:"path"`
	TLS         string `json:"tls"`
	SNI         string `json:"sni"`
	ALPN        string `json:"alpn"`
	Fingerprint string `json:"fp"`
}

func ParseLink(raw string) (Link, error) {
	switch {
	case strings.HasPrefix(raw, "vless://"):
		return parseVLESS(raw)
	case strings.HasPrefix(raw, "vmess://"):
		return parseVMess(raw)
	default:
		return Link{}, ErrUnsupportedProtocol
	}
}

func parseVLESS(raw string) (Link, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return Link{}, fmt.Errorf("parse vless url: %w", err)
	}
	if parsed.User == nil {
		return Link{}, errors.New("missing uuid")
	}
	uuid := parsed.User.Username()
	host, port, err := splitHostPort(parsed.Host, 443)
	if err != nil {
		return Link{}, err
	}

	query := parsed.Query()
	return Link{
		Protocol:      "vless",
		Name:          urlDecodeFragment(parsed.Fragment),
		Address:       host,
		Port:          port,
		UUID:          uuid,
		Encryption:    query.Get("encryption"),
		Security:      query.Get("security"),
		Transport:     firstNonEmpty(query.Get("type"), query.Get("transport"), "tcp"),
		SNI:           query.Get("sni"),
		Host:          query.Get("host"),
		Path:          query.Get("path"),
		Fingerprint:   query.Get("fp"),
		Flow:          query.Get("flow"),
		ALPN:          splitCSV(query.Get("alpn")),
		ServiceName:   query.Get("serviceName"),
		AllowInsecure: query.Get("allowInsecure") == "1",
		Raw:           raw,
	}, nil
}

func parseVMess(raw string) (Link, error) {
	encoded := strings.TrimPrefix(raw, "vmess://")
	payloadBytes, err := decodeBase64(encoded)
	if err != nil {
		return Link{}, fmt.Errorf("decode vmess payload: %w", err)
	}

	var payload vmessPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return Link{}, fmt.Errorf("unmarshal vmess payload: %w", err)
	}

	port, err := strconv.Atoi(payload.Port)
	if err != nil {
		return Link{}, fmt.Errorf("parse vmess port: %w", err)
	}

	return Link{
		Protocol:    "vmess",
		Name:        payload.Name,
		Address:     payload.Address,
		Port:        port,
		UUID:        payload.ID,
		Encryption:  payload.AlterID,
		Security:    payload.TLS,
		Transport:   firstNonEmpty(payload.Network, "tcp"),
		SNI:         payload.SNI,
		Host:        payload.Host,
		Path:        payload.Path,
		Fingerprint: payload.Fingerprint,
		ALPN:        splitCSV(payload.ALPN),
		Raw:         raw,
	}, nil
}

func decodeBase64(data string) ([]byte, error) {
	trimmed := strings.TrimSpace(data)
	if trimmed == "" {
		return nil, errors.New("empty base64 data")
	}
	for _, encoding := range []*base64.Encoding{
		base64.RawStdEncoding,
		base64.StdEncoding,
		base64.RawURLEncoding,
		base64.URLEncoding,
	} {
		if decoded, err := encoding.DecodeString(trimmed); err == nil {
			return decoded, nil
		}
	}
	return nil, errors.New("invalid base64 data")
}

func splitHostPort(value string, defaultPort int) (string, int, error) {
	if value == "" {
		return "", 0, errors.New("missing host")
	}
	host, portStr, err := net.SplitHostPort(value)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") {
			return value, defaultPort, nil
		}
		return "", 0, fmt.Errorf("split host/port: %w", err)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return "", 0, fmt.Errorf("parse port: %w", err)
	}

	return host, port, nil
}

func splitCSV(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item != "" {
			out = append(out, item)
		}
	}
	return out
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func urlDecodeFragment(value string) string {
	if value == "" {
		return ""
	}
	decoded, err := url.PathUnescape(value)
	if err != nil {
		return value
	}
	return decoded
}
