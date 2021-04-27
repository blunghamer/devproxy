package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "devproxy",
	Short: "devproxy",
	Long:  `fast and transparent proxy to avoid copying proxy credentials into each app`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}

var config DevProxyConfig
var bindto string

func init() {
	rootCmd.PersistentFlags().StringVarP(&bindto, bindToKey, "b", "0.0.0.0:3128", "target bound interface to run proxy on")
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	toolname := "devproxy"
	viper.SetConfigName(toolname)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath(fmt.Sprintf("/etc/%v/", toolname))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.%v", toolname))
	viper.AddConfigPath(fmt.Sprintf("$HOME/.config/%v", toolname))

	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("fatal error in readinconfig: %s, config %v", err, viper.ConfigFileUsed())
	}

	log.Println("Conf file read", viper.AllKeys())
	err = viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("fatal error unmarshalling config file: %s", err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Unable to find home directory", err)
	}

	keystorefile := fmt.Sprintf("%v/.%v.yaml", home, toolname)

	if _, err := os.Stat(keystorefile); os.IsNotExist(err) {
		log.Println("Devproxy key file does not exist, please run devproxy cred before running")
	} else {
		password, err := getCredentials(config.Appname, config.Proxyuser)
		if err != nil {
			log.Fatalf("Unable to get proxy credentials for user %v error %v", config.Proxyuser, err)
		}
		config.Proxypassword = password
	}
}

// Execute root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
