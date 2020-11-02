// +build windows

package cmd

import (
	"bytes"
	"fmt"
	"log"
	"syscall"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/ssh/terminal"
)

func init() {
	var appname string
	var proxyUser string
	var password string

	rootCmd.AddCommand(creds)
	creds.Flags().StringVarP(&appname, appNameKey, "n", "dev-proxy", "proxy service appname")
	creds.Flags().StringVarP(&proxyUser, proxyUserKey, "u", "", "proxyuser name")
	creds.Flags().StringVarP(&password, passwordKey, "p", "", "password")

	viper.BindPFlag(appNameKey, creds.Flags().Lookup(appNameKey))
	viper.BindPFlag(proxyUserKey, creds.Flags().Lookup(proxyUserKey))
	viper.BindPFlag(passwordKey, creds.Flags().Lookup(passwordKey))
}

var creds = &cobra.Command{
	Use:   "cred",
	Short: "set credentials",
	Long:  `set credentials to operating system secret store`,
	Run: func(cmd *cobra.Command, args []string) {
		// log.Print(viper.AllSettings())

		if viper.GetString(passwordKey) == "" {
			fmt.Print("Password: ")
			bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				log.Fatal("Unable to read password")
			}
			viper.Set(passwordKey, string(bytePassword))
		}

		if viper.GetString(proxyUserKey) == "" {
			log.Fatal("Unable to get current user, please supply on command line")
		}

		err := setCredentials(viper.GetString(appNameKey), viper.GetString(proxyUserKey), viper.GetString(passwordKey))
		if err != nil {
			log.Fatal("Unable to set credentials to keystore")
		}
	},
}

// set password
func setCredentials(service string, username string, password string) error {
	log.Println("Setting", service, "credentials for user", username)
	err := keyring.Set(service, username, password)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// get password windows returns utf16 (uhhh) and we have to convert his to utf8 runes
func getCredentials(service string, username string) (string, error) {
	log.Println("Getting", service, "credentials for user", username)
	secret, err := keyring.Get(service, username)
	if err != nil {
		log.Println(err)
		return "", err
	}

	// log.Println(secret)

	/*
		res, err := decodeUTF16([]byte(secret))
		if err != nil {
			log.Println(err)
			return "", err
		}
	*/

	return secret, nil
}

func decodeUTF16(b []byte) (string, error) {

	if len(b)%2 != 0 {
		return "", fmt.Errorf("Must have even length byte slice")
	}

	u16s := make([]uint16, 1)

	ret := &bytes.Buffer{}

	b8buf := make([]byte, 4)

	lb := len(b)
	for i := 0; i < lb; i += 2 {
		u16s[0] = uint16(b[i]) + (uint16(b[i+1]) << 8)
		r := utf16.Decode(u16s)
		n := utf8.EncodeRune(b8buf, r[0])
		ret.Write(b8buf[:n])
	}

	return ret.String(), nil
}
