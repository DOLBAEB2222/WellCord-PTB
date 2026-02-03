package main

import "strings"

type ProxyConfig struct {
	Mode   string `yaml:"mode" mapstructure:"mode"`
	Server string `yaml:"server" mapstructure:"server"`
	Bypass string `yaml:"bypass" mapstructure:"bypass"`
	PACURL string `yaml:"pac_url" mapstructure:"pac_url"`
}

func applyProxyArgs(proxy ProxyConfig, args *[]string) {
	mode := strings.ToLower(strings.TrimSpace(proxy.Mode))
	switch mode {
	case "", "system":
		return
	case "direct", "none", "off":
		*args = append(*args, "--no-proxy-server")
	case "pac":
		if strings.TrimSpace(proxy.PACURL) == "" {
			return
		}
		*args = append(*args, "--proxy-pac-url="+proxy.PACURL)
	case "fixed", "fixed_servers":
		if strings.TrimSpace(proxy.Server) == "" {
			return
		}
		*args = append(*args, "--proxy-server="+proxy.Server)
	default:
		return
	}

	if strings.TrimSpace(proxy.Bypass) != "" {
		*args = append(*args, "--proxy-bypass-list="+proxy.Bypass)
	}
}
