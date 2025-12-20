package core

// CertConfig contains certificate configuration
type CertConfig struct {
	Enabled           bool
	JDKSecurityPath   string
	SystemCacertsPath string
}

// DownloadOptions contains options for download and installation
type DownloadOptions struct {
	DownloadURL   string
	CacheDir      string
	InstallPath   string
	SDKType       string
	Distribution  string
	Version       string
	KeepCache     bool
	CertConfig    CertConfig
	Username      string // HTTP Basic Auth username (optional)
	Password      string // HTTP Basic Auth password (optional)
}
