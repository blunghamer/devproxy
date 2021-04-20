package cmd

import (
	"log"
	"net/http"

	"github.com/elazarl/goproxy"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	rootCmd.AddCommand(runDirect)
}

// NewDirectProxyHTTPServer customizes behaviour
// we do not take into account the environment variables
func NewDirectProxyHTTPServer() *goproxy.ProxyHttpServer {
	proxy := goproxy.NewProxyHttpServer()
	proxy.Tr.DialContext = nil
	proxy.Tr.Proxy = nil
	proxy.ConnectDial = nil
	proxy.Verbose = true
	return proxy
}

var runDirect = &cobra.Command{
	Use:   "direct",
	Short: "direct proxy to internet",
	Long:  `direct proxy to internet`,
	Run: func(cmd *cobra.Command, args []string) {
		// log.Print(viper.AllSettings())
		endProxy := NewDirectProxyHTTPServer()
		log.Println("serving direct proxy server at", viper.GetString(bindToKey))
		log.Fatal(http.ListenAndServe(viper.GetString(bindToKey), endProxy))
	},
}
