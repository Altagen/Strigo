package downloader

// CertConfig contains certificate configuration
type CertConfig struct {
	// Enabled indicates if certificate configuration is enabled
	Enabled bool

	// JDKSecurityPath is the relative path to the cacerts file in the JDK
	// For example: "lib/security/cacerts"
	JDKSecurityPath string

	// SystemCacertsPath is the absolute path to system certificates
	// For example: "/etc/ssl/certs/java/cacerts"
	SystemCacertsPath string
}
