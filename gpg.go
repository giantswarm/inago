package infraconfigparser

import (
	"bytes"
	"io/ioutil"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
)

func decryptGPGBytes(raw []byte) ([]byte, error) {
	pass, err := gpgPass()
	if err != nil {
		return nil, maskAny(err)
	}

	return decryptGPGBytesWithPass(raw, pass)
}

const (
	// TODO here we need to make the config dir configurable
	homeConfigDirName = ".giantswarm/releaseit/"
)

func gpgPass() ([]byte, error) {
	home, err := homedir.Dir()
	if err != nil {
		return nil, maskAny(err)
	}

	// TODO here we need to make the gpg pass configurable
	return ioutil.ReadFile(filepath.Join(home, homeConfigDirName, ".gpgpass"))
}

func decryptGPGBytesWithPass(raw, pass []byte) ([]byte, error) {
	decbuf := bytes.NewBuffer(raw)
	result, err := armor.Decode(decbuf)
	if err != nil {
		return nil, maskAny(err)
	}

	md, err := openpgp.ReadMessage(result.Body, nil, func(keys []openpgp.Key, symmetric bool) ([]byte, error) {
		return pass, nil
	}, nil)
	if err != nil {
		return nil, maskAny(err)
	}

	b, err := ioutil.ReadAll(md.UnverifiedBody)
	if err != nil {
		return nil, maskAny(err)
	}

	return b, nil
}
