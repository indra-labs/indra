package storage

import (
	"github.com/spf13/viper"
	"os"
	"strings"
)

var (
	newKeyGenerated bool
	isNewDB         bool
	noKeyProvided   bool
	key             Key
)

func configure() {

	log.I.Ln("initializing storage")

	configureDirPath()
	configureFile()
	configureKey()
}

func configureKey() {

	log.I.Ln("looking for encryption key")

	var err error

	if viper.GetString(storeKeyFlag) != "" {

		log.I.Ln("found key")

		key.Decode(viper.GetString(storeKeyFlag))

		return
	}

	log.I.Ln("no key found, checking for keyfile")

	if viper.GetString(storeKeyFileFlag) != "" {

		var fileInfo os.FileInfo

		if fileInfo, err = os.Stat(viper.GetString(storeKeyFileFlag)); err != nil {
			startupErrors <- err
			return
		}

		log.I.Ln("keyfile found")

		if fileInfo.Mode() != 0600 {
			log.W.Ln("keyfile permissions are too open:", fileInfo.Mode())
			log.W.Ln("It is recommended that you change them to 0600")
		}

		var keyBytes []byte

		if keyBytes, err = os.ReadFile(viper.GetString(storeKeyFileFlag)); err != nil {
			startupErrors <- err
			return
		}

		key.Decode(string(keyBytes))

		return
	}

	if !isNewDB {

		log.I.Ln("no keyfile found")

		noKeyProvided = true

		return
	}

	log.I.Ln("no keyfile found, generating a new key")

	if key, err = KeyGen(); err != nil {
		startupErrors <- err
		return
	}

	log.W.Ln(`
--------------------------------------------------------
--
-- WARNING: The following key will be used to store
-- your database securely, please ensure that you make
-- a copy and store it in a secure place before using
-- this software in a production environment.
--
--
-- Failure to store this key properly will result in
-- no longer being able to decrypt this database.
--
--
-- It is recommended to use the following to generate
-- your key:
--
--     indra seed keygen
--
-- OR
--
--     indra seed keygen --keyfile=/path/to/keyfile
--
--
-- YOU HAVE BEEN WARNED!
--
-------------------------------------------------------
-- KEY:`, key.Encode(), `--
-------------------------------------------------------
`)

	newKeyGenerated = true

	viper.Set(storeKeyFlag, key.Encode())
}

func configureDirPath() {

	var err error

	if viper.GetString(storeFilePathFlag) == "" {
		viper.Set(storeFilePathFlag, viper.GetString("data-dir")+"/"+fileName)
	}

	err = os.MkdirAll(
		strings.TrimSuffix(viper.GetString(storeFilePathFlag), "/"+fileName),
		0755,
	)

	if err != nil {
		startupErrors <- err
		return
	}

}

func configureFile() {

	log.I.Ln("using storage db path:")
	log.I.Ln("-", viper.GetString(storeFilePathFlag))

	log.I.Ln("checking if database exists")

	var err error

	if _, err = os.Stat(viper.GetString(storeFilePathFlag)); err != nil {

		log.I.Ln("no database found, creating a new one")

		isNewDB = true

		return
	}

	log.I.Ln("database found")
}
