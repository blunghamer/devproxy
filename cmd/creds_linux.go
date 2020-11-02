// +build linux

package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"syscall"

	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
	"golang.org/x/crypto/ssh/terminal"

	"gopkg.in/yaml.v2"
)

func init() {
	var appname string
	var proxyUser string
	var password string
	var useKeyring bool

	rootCmd.AddCommand(creds)
	creds.Flags().StringVarP(&appname, "appname", "n", "dev-proxy", "proxy service appname")
	creds.Flags().StringVarP(&proxyUser, "proxyuser", "u", "", "proxyuser name")
	creds.Flags().StringVarP(&password, "password", "p", "", "password")
	creds.Flags().BoolVarP(&useKeyring, "keyring", "k", false, "if gnome dbus keyring interface is available and should used for credentials")

	viper.BindPFlag("appname", creds.Flags().Lookup("appname"))
	viper.BindPFlag("proxyuser", creds.Flags().Lookup("proxyuser"))
	viper.BindPFlag("password", creds.Flags().Lookup("password"))
	viper.BindPFlag("keyring", creds.Flags().Lookup("keyring"))

	// on headless linux, just use a plain in memory store
	// if keyring should not be used
	if !useKeyring {
		keyring.MockInit()
	}
}

var creds = &cobra.Command{
	Use:   "cred",
	Short: "set credentials",
	Long:  `set credentials to operating system secret store`,
	Run: func(cmd *cobra.Command, args []string) {
		// log.Print(viper.AllSettings())

		if viper.GetString("password") == "" {
			fmt.Print("Password: ")
			bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
			if err != nil {
				log.Fatal("Unable to read password")
			}
			viper.Set("password", string(bytePassword))
		}

		if viper.GetString("proxyuser") == "" {
			log.Fatal("Unable to get current user, please supply on command line")
		}

		err := setCredentials(viper.GetString("appname"), viper.GetString("proxyuser"), viper.GetString("password"))
		if err != nil {
			log.Fatal("Unable to set credentials to keystore")
		}
	},
}

// set password
func setCredentials(service string, username string, password string) error {
	if viper.GetBool("keyring") {
		log.Println("Setting", service, "credentials for user", username)
		err := keyring.Set(service, username, password)
		if err != nil {
			log.Println(err)
			return err
		}
		return nil
	}
	log.Println("Setting", service, "credentials for user", username)
	EncryptGCM([]byte(password))
	// Read back
	/*
		res, err := DecryptGCM()
		if err != nil {
			log.Println("Unable to decode payload", err)
		}
		log.Println("Result is", res)
	*/
	return nil
}

// get password windows returns utf16 (uhhh) and we have to convert his to utf8 runes
func getCredentials(service string, username string) (string, error) {
	if viper.GetBool("keyring") {
		log.Println("Getting", service, "credentials for user", username)
		secret, err := keyring.Get(service, username)
		if err != nil {
			log.Println("Error retrieving secret", err)
			return "", err
		}
		return secret, nil
	}
	res, err := DecryptGCM()
	return res, err
}

// EncryptGCM encrypt with AES / Galois Counter Mode taken and extended from https://golang.org/src/crypto/cipher/example_test.go
func EncryptGCM(plaintext []byte) error {

	home, err := os.UserHomeDir()
	if err != nil {
		log.Println("Unable to find home directory", err)
		return err
	}

	keystorefile := home + "/.devproxy.yaml"

	if _, err := os.Stat(keystorefile); os.IsNotExist(err) {
		log.Println("Devproxy key file does not exists generating...")

		key := make([]byte, 32)
		_, err = rand.Read(key)
		if err != nil {
			log.Println("unable to create random key material", err)
			return err
		}

		// Never use more than 2^32 random nonces with a given key because of the risk of a repeat.
		nonce := make([]byte, 12)
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			log.Println("unable to create nonce", err)
			return err
		}

		outb, err := yaml.Marshal(&KeyStore{Key: hex.EncodeToString(key), Nonce: hex.EncodeToString(nonce)})
		if err != nil {
			log.Println("unable to mashal keystore", err)
			return err
		}

		err = ioutil.WriteFile(keystorefile, outb, 0640)
		if err != nil {
			log.Println("unable to write keystore file", err)
			return err
		}
	}

	inby, err := ioutil.ReadFile(keystorefile)
	if err != nil {
		log.Println("Unable to read keystore file", err)
		return err
	}

	var ks KeyStore
	err = yaml.Unmarshal(inby, &ks)
	if err != nil {
		log.Println("Unable to unmarshal corrupt keystore file", err)
		return err
	}

	key, _ := hex.DecodeString(ks.Key)
	nonce, _ := hex.DecodeString(ks.Nonce)

	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	ciphertext := aesgcm.Seal(nil, nonce, plaintext, nil)
	ks.Payload = hex.EncodeToString(ciphertext)

	outb, err := yaml.Marshal(&ks)
	if err != nil {
		log.Println("unable to mashal keystore", err)
		return err
	}

	err = ioutil.WriteFile(keystorefile, outb, 0640)
	if err != nil {
		log.Println("unable to write keystore file", err)
		return err
	}

	return nil
}

// DecryptGCM decrypt with AES / Galois Counter Mode taken and extended from https://golang.org/src/crypto/cipher/example_test.go
func DecryptGCM() (string, error) {

	home, err := os.UserHomeDir()
	if err != nil {
		log.Println("Unable to find home directory", err)
		return "", err
	}

	keystorefile := home + "/.devproxy.yaml"
	if _, err := os.Stat(keystorefile); os.IsNotExist(err) {
		log.Println("Devproxy key file does not exists bailing out")
		return "", err
	}

	inby, err := ioutil.ReadFile(keystorefile)
	if err != nil {
		log.Println("Unable to read keystore file", err)
		return "", err
	}

	var ks KeyStore
	err = yaml.Unmarshal(inby, &ks)
	if err != nil {
		log.Println("Unable to unmarshal corrupt keystore file", err)
		return "", err
	}

	key, _ := hex.DecodeString(ks.Key)
	nonce, _ := hex.DecodeString(ks.Nonce)
	ciphertext, _ := hex.DecodeString(ks.Payload)

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
