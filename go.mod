module github.com/blunghamer/devproxy

go 1.15

replace github.com/elazarl/goproxy => github.com/blunghamer/goproxy v0.0.0-20200829102833-ad1ea8cd5e16

require (
	github.com/elazarl/goproxy v0.0.0-20201021153353-00ad82a08272
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/zalando/go-keyring v0.1.0
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897
	gopkg.in/yaml.v2 v2.3.0
)
