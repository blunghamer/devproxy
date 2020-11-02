package cmd

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/elazarl/goproxy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	var (
		appname    string
		proxyUser  string
		password   string
		httpProxy  string
		httpsProxy string
	)

	runProxyChain.Flags().StringVarP(&appname, appNameKey, "n", "dev-proxy", "proxy service appname")
	runProxyChain.Flags().StringVarP(&proxyUser, proxyUserKey, "u", "", "proxyuser name")
	runProxyChain.Flags().StringVarP(&password, passwordKey, "p", "", "password")
	runProxyChain.Flags().StringVarP(&httpProxy, httpProxyKey, "r", "", "http proxy address")
	runProxyChain.Flags().StringVarP(&httpsProxy, httpsProxyKey, "s", "", "https proxy address")

	// viper.BindPFlag(bindToKey, runProxyChain.Flags().Lookup(bindToKey))
	viper.BindPFlag(appNameKey, runProxyChain.Flags().Lookup(appNameKey))
	viper.BindPFlag(passwordKey, runProxyChain.Flags().Lookup(passwordKey))
	viper.BindPFlag(proxyUserKey, runProxyChain.Flags().Lookup(proxyUserKey))
	viper.BindPFlag(httpProxyKey, runProxyChain.Flags().Lookup(httpProxyKey))
	viper.BindPFlag(httpsProxyKey, runProxyChain.Flags().Lookup(httpsProxyKey))

	rootCmd.AddCommand(runProxyChain)
}

var runProxyChain = &cobra.Command{
	Use:   "chain",
	Short: "chain devproxy against corporate proxy",
	Long:  `chain devproxy against corporate proxy`,
	Run: func(cmd *cobra.Command, args []string) {

		bindTo := viper.GetString(bindToKey)
		password := viper.GetString(passwordKey)
		appname := viper.GetString(appNameKey)
		proxyUser := viper.GetString(proxyUserKey)
		httpProxy := viper.GetString(httpProxyKey)
		httpsProxy := viper.GetString(httpsProxyKey)

		// get password
		if password == "" {
			var err error
			password, err = getCredentials(appname, proxyUser)
			if err != nil {
				log.Fatalf("Unable to get proxy credentials for user %v error %v", proxyUser, err)
			}
		}

		// start middle proxy with password inject into connections
		middleProxy := goproxy.NewProxyHttpServer()
		middleProxy.Verbose = true
		middleProxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
			return url.Parse(httpProxy)
		}
		connectReqHandler := func(req *http.Request) {
			SetBasicAuth(proxyUser, password, req)
		}
		middleProxy.ConnectDial = middleProxy.NewConnectDialToProxyWithHandler(httpsProxy, connectReqHandler)
		middleProxy.OnRequest().DoLate(goproxy.FuncReqHandler(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			SetBasicAuth(proxyUser, password, req)
			return req, nil
		}))

		log.Println("serving development proxy server at", bindTo, "user", proxyUser)
		log.Println("forwarding traffic to", httpProxy, httpsProxy)
		log.Fatal(http.ListenAndServe(bindTo, middleProxy))
		fmt.Println("Devproxy")
	},
}

const (
	// ProxyAuthHeader contains standard header to set basic auth for proxy access
	ProxyAuthHeader = "Proxy-Authorization"
)

// SetBasicAuth add basic auth for proxy access to the request
func SetBasicAuth(username, password string, req *http.Request) {
	req.Header.Add(ProxyAuthHeader, fmt.Sprintf("Basic %s", basicAuth(username, password)))
}

func basicAuth(username, password string) string {
	return base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
}

// GetBasicAuth read basic auth header
func GetBasicAuth(req *http.Request) (username, password string, ok bool) {
	auth := req.Header.Get(ProxyAuthHeader)
	if auth == "" {
		return
	}

	const prefix = "Basic "
	if !strings.HasPrefix(auth, prefix) {
		return
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return
	}
	return cs[:s], cs[s+1:], true
}
