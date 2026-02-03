package vpn

type RoutingConfig struct {
	DomainStrategy string        `json:"domainStrategy,omitempty"`
	Rules          []RoutingRule `json:"rules"`
}

type RoutingRule struct {
	Type     string   `json:"type"`
	Domain   []string `json:"domain,omitempty"`
	IP       []string `json:"ip,omitempty"`
	Port     string   `json:"port,omitempty"`
	Network  string   `json:"network,omitempty"`
	Outbound string   `json:"outboundTag"`
}

func DefaultRouting() *RoutingConfig {
	return &RoutingConfig{
		DomainStrategy: "AsIs",
		Rules: []RoutingRule{
			{
				Type:     "field",
				Domain:   []string{"discord.com", "discord.gg"},
				Outbound: "proxy",
			},
			{
				Type:     "field",
				Domain:   []string{"yandex.ru", "vk.com"},
				Outbound: "direct",
			},
			{
				Type:     "field",
				IP:       []string{"geoip:private"},
				Outbound: "direct",
			},
		},
	}
}

func WithOutboundTags(config XrayConfig, proxyTag string) XrayConfig {
	if config.Routing == nil {
		config.Routing = DefaultRouting()
	}
	for idx, rule := range config.Routing.Rules {
		if rule.Outbound == "proxy" {
			rule.Outbound = proxyTag
			config.Routing.Rules[idx] = rule
		}
	}
	return config
}
