package vpn

import (
	"errors"
)

type XrayConfig struct {
	Log       LogConfig        `json:"log,omitempty"`
	Inbounds  []InboundConfig  `json:"inbounds"`
	Outbounds []OutboundConfig `json:"outbounds"`
	Routing   *RoutingConfig   `json:"routing,omitempty"`
}

type LogConfig struct {
	LogLevel string `json:"loglevel,omitempty"`
}

type InboundConfig struct {
	Port     int               `json:"port"`
	Listen   string            `json:"listen"`
	Protocol string            `json:"protocol"`
	Settings map[string]string `json:"settings,omitempty"`
}

type OutboundConfig struct {
	Protocol       string                 `json:"protocol"`
	Settings       map[string]interface{} `json:"settings"`
	StreamSettings StreamSettings         `json:"streamSettings,omitempty"`
	Tag            string                 `json:"tag,omitempty"`
}

type StreamSettings struct {
	Network         string             `json:"network,omitempty"`
	Security        string             `json:"security,omitempty"`
	TLSSettings     *TLSSettings       `json:"tlsSettings,omitempty"`
	RealitySettings *RealitySettings   `json:"realitySettings,omitempty"`
	WSSettings      *WebSocketSettings `json:"wsSettings,omitempty"`
	GRPCSettings    *GRPCSettings      `json:"grpcSettings,omitempty"`
	HTTPSettings    *HTTPSettings      `json:"httpSettings,omitempty"`
}

type TLSSettings struct {
	ServerName    string   `json:"serverName,omitempty"`
	AllowInsecure bool     `json:"allowInsecure,omitempty"`
	ALPN          []string `json:"alpn,omitempty"`
	Fingerprint   string   `json:"fingerprint,omitempty"`
}

type RealitySettings struct {
	ServerName  string `json:"serverName,omitempty"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

type WebSocketSettings struct {
	Path    string            `json:"path,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type GRPCSettings struct {
	ServiceName string `json:"serviceName,omitempty"`
}

type HTTPSettings struct {
	Path []string `json:"path,omitempty"`
	Host []string `json:"host,omitempty"`
}

func BuildXrayConfig(link Link) (XrayConfig, error) {
	if link.Address == "" || link.Port == 0 || link.UUID == "" {
		return XrayConfig{}, errors.New("missing required link fields")
	}

	outbound, err := buildOutbound(link)
	if err != nil {
		return XrayConfig{}, err
	}

	config := XrayConfig{
		Log: LogConfig{
			LogLevel: "warning",
		},
		Inbounds: []InboundConfig{
			{
				Port:     10808,
				Listen:   "127.0.0.1",
				Protocol: "socks",
				Settings: map[string]string{
					"udp": "true",
				},
			},
		},
		Outbounds: []OutboundConfig{outbound, {
			Protocol: "freedom",
			Tag:      "direct",
			Settings: map[string]interface{}{},
		}},
		Routing: DefaultRouting(),
	}

	config = WithOutboundTags(config, "proxy")

	return config, nil
}

func buildOutbound(link Link) (OutboundConfig, error) {
	switch link.Protocol {
	case "vless":
		return buildVLESSOutbound(link), nil
	case "vmess":
		return buildVMessOutbound(link), nil
	default:
		return OutboundConfig{}, ErrUnsupportedProtocol
	}
}

func buildVLESSOutbound(link Link) OutboundConfig {
	settings := map[string]interface{}{
		"vnext": []map[string]interface{}{
			{
				"address": link.Address,
				"port":    link.Port,
				"users": []map[string]interface{}{
					{
						"id":         link.UUID,
						"encryption": firstNonEmpty(link.Encryption, "none"),
						"flow":       link.Flow,
					},
				},
			},
		},
	}

	return OutboundConfig{
		Protocol:       "vless",
		Settings:       settings,
		StreamSettings: buildStreamSettings(link),
		Tag:            "proxy",
	}
}

func buildVMessOutbound(link Link) OutboundConfig {
	settings := map[string]interface{}{
		"vnext": []map[string]interface{}{
			{
				"address": link.Address,
				"port":    link.Port,
				"users": []map[string]interface{}{
					{
						"id":       link.UUID,
						"alterId":  link.Encryption,
						"security": "auto",
					},
				},
			},
		},
	}

	return OutboundConfig{
		Protocol:       "vmess",
		Settings:       settings,
		StreamSettings: buildStreamSettings(link),
		Tag:            "proxy",
	}
}

func buildStreamSettings(link Link) StreamSettings {
	settings := StreamSettings{
		Network: firstNonEmpty(link.Transport, "tcp"),
	}

	security := normalizeSecurity(link.Security)
	if security != "" {
		settings.Security = security
	}

	switch settings.Network {
	case "ws":
		wsHeaders := map[string]string{}
		if link.Host != "" {
			wsHeaders["Host"] = link.Host
		}
		settings.WSSettings = &WebSocketSettings{
			Path:    link.Path,
			Headers: wsHeaders,
		}
	case "grpc":
		settings.GRPCSettings = &GRPCSettings{
			ServiceName: link.ServiceName,
		}
	case "h2", "http", "http2":
		settings.HTTPSettings = &HTTPSettings{
			Path: splitCSV(link.Path),
			Host: splitCSV(link.Host),
		}
	}

	if security == "tls" {
		settings.TLSSettings = buildTLSSettings(link)
	}
	if security == "reality" {
		settings.RealitySettings = &RealitySettings{
			ServerName:  link.SNI,
			Fingerprint: link.Fingerprint,
		}
	}

	return settings
}

func buildTLSSettings(link Link) *TLSSettings {
	if link.SNI == "" && len(link.ALPN) == 0 && link.Fingerprint == "" && !link.AllowInsecure {
		return nil
	}
	return &TLSSettings{
		ServerName:    link.SNI,
		AllowInsecure: link.AllowInsecure,
		ALPN:          link.ALPN,
		Fingerprint:   link.Fingerprint,
	}
}

func normalizeSecurity(value string) string {
	switch value {
	case "tls", "xtls":
		return "tls"
	case "reality":
		return "reality"
	default:
		return ""
	}
}
