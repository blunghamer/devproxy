package cmd

// config keys for static checking of config file, command line flags, environment variable flags
const bindToKey = "bindto"
const proxyUserKey = "proxyuser"
const appNameKey = "dev-proxy"
const passwordKey = "password"
const httpProxyKey = "httpproxy"
const httpsProxyKey = "httpsproxy"

const httpTest = "httpTest"
const httpsTest = "httpsTest"

// DevProxyConfig for the proxy
type DevProxyConfig struct {
	Bindto        string
	Appname       string
	Proxyuser     string
	Proxypassword string // it is not ok to place this in the config file use env variables or command line instread
	HTTPProxy     string
	HTTPSProxy    string

	HTTPSTest string
	HTTPTest  string
}

// KeyStore enables persistence of nonce and key for Aes GCM
type KeyStore struct {
	Key     string
	Nonce   string
	Payload string
}
