module github.com/blunghamer/devproxy

go 1.15

replace github.com/elazarl/goproxy => github.com/blunghamer/goproxy v0.0.0-20201107215641-a0ef4c6459e2

require (
	github.com/alexellis/go-execute v0.0.0-20201205082949-69a2cde04f4f // indirect
	github.com/elazarl/goproxy v0.0.0-20201021153353-00ad82a08272
	github.com/spf13/cobra v1.1.1
	github.com/spf13/viper v1.7.1
	github.com/zalando/go-keyring v0.1.0
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897
	gopkg.in/yaml.v2 v2.3.0
)
