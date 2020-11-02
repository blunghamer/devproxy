package cmd

import (
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(runAuto)
}

var runAuto = &cobra.Command{
	Use:   "auto",
	Short: "run auto proxy to internet",
	Long:  `run auto proxy to internet`,
	Run: func(cmd *cobra.Command, args []string) {
		ap := NewAutoProxy()
		ap.run()
	},
}

// AutoProxy can switch config on its own
type AutoProxy struct {
	direct  bool
	Direct  *goproxy.ProxyHttpServer
	Chained *goproxy.ProxyHttpServer
}

// NewAutoProxy creates an AutoProxy
func NewAutoProxy() *AutoProxy {
	ap := &AutoProxy{direct: true}
	ap.Direct = NewDirectProxy(ap.reconfigure)
	ap.Chained = NewChainedProxy(ap.reconfigure)
	return ap
}

func (a *AutoProxy) reconfigure(err error) {
	log.Println("There seems to be an error", err)
	if strings.Contains(err.Error(), "timeout") {
		_, err := http.Get(config.HTTPTest)
		if err != nil {
			log.Println("Got timeout need to reconfigure")
		} else {
			log.Println("Was not timeout to proxy server")
		}
	}
}

func (a *AutoProxy) tryDirect() bool {
	cli := http.Client{Transport: &http.Transport{
		Proxy: nil,
		DialContext: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 10 * time.Second,
			DualStack: true,
		}).DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	},
	}
	rp, err := cli.Get(config.HTTPTest)
	if err != nil {
		log.Println("no direct connect possible", err)
		return false
	}
	log.Println("Direct HTTP Get out succes", rp.StatusCode)
	return true
}

func (a *AutoProxy) run() {
	if a.tryDirect() {
		log.Println("serving direct proxy server at", config.Bindto)
		log.Fatal(http.ListenAndServe(config.Bindto, a.Direct))
	} else {
		log.Println("serving chained proxy server at", config.Bindto)
		log.Fatal(http.ListenAndServe(config.Bindto, a.Chained))
	}
}

// NewDirectProxy explicitly bypasses env variables and direclty dials out
func NewDirectProxy(fun goproxy.ErrorHook) *goproxy.ProxyHttpServer {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Tr.DialContext = nil
	proxy.Tr.Proxy = nil
	proxy.ConnectDial = nil
	proxy.Verbose = true
	proxy.Hook = fun
	return proxy
}

// NewChainedProxy dials out to chained proxy server
func NewChainedProxy(fun goproxy.ErrorHook) *goproxy.ProxyHttpServer {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxy.Hook = fun
	proxy.Tr.Proxy = func(req *http.Request) (*url.URL, error) {
		return url.Parse(config.HTTPProxy)
	}
	connectReqHandler := func(req *http.Request) {
		SetBasicAuth(config.Proxyuser, config.Proxypassword, req)
	}
	proxy.ConnectDial = proxy.NewConnectDialToProxyWithHandler(config.HTTPSProxy, connectReqHandler)
	proxy.OnRequest().DoLate(goproxy.FuncReqHandler(func(req *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		SetBasicAuth(config.Proxyuser, config.Proxypassword, req)
		return req, nil
	}))
	return proxy
}
