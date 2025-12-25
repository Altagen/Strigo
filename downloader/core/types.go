package core

// DownloadOptions contains options for download and installation
type DownloadOptions struct {
	DownloadURL  string
	CacheDir     string
	InstallPath  string
	SDKType      string
	Distribution string
	Version      string
	KeepCache    bool
	Username     string // HTTP Basic Auth username (optional)
	Password     string // HTTP Basic Auth password (optional)
}
