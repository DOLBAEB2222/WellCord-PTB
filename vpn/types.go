package vpn

type Link struct {
	Protocol      string
	Name          string
	Address       string
	Port          int
	UUID          string
	Encryption    string
	Security      string
	Transport     string
	SNI           string
	Host          string
	Path          string
	Fingerprint   string
	Flow          string
	ALPN          []string
	ServiceName   string
	AllowInsecure bool
	Raw           string
}

type ImportResult struct {
	Links  []Link
	Errors []error
}
