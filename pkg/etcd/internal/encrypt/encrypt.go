package encrypt

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/giantswarm/microerror"
	"golang.org/x/crypto/openpgp"
)

// Encrypt data with passphrase.
func data(value []byte, pass string) (ciphertext []byte, err error) {
	buf := bytes.NewBuffer(nil)

	encrypter, err := openpgp.SymmetricallyEncrypt(buf, []byte(pass), nil, nil)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	_, err = encrypter.Write(value)
	if err != nil {
		return nil, microerror.Mask(err)
	}

	encrypter.Close()

	return buf.Bytes(), nil
}

// Encrypts file from srcPath and writes encrypted data to dstPart.
func File(srcPath string, dstPart string, passphrase string) error {
	contents, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return microerror.Mask(err)
	}

	encData, err := data(contents, passphrase)
	if err != nil {
		return microerror.Mask(err)
	}

	err = ioutil.WriteFile(dstPart, encData, os.FileMode(0600))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
