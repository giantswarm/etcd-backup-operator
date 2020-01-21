package backup

import (
	"bytes"
	"fmt"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"golang.org/x/crypto/openpgp"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)

const (
	etcdctlCmd = "etcdctl"
	awsCmd     = "Aws"
	tgzExt     = ".tar.gz"
	encExt     = ".enc"
	dbExt      = ".db"
)

// Outputs timestamp.
func getTimeStamp() string {
	return time.Now().Format("2006-01-02T15-04-05")
}

// Executes command and outputs stdout+stderr and error if any.
// Arguments:
// - cmd  - command to execute
// - args - arguments for command
// - envs - envronment variables
func execCmd(cmd string, args []string, envs []string, logger micrologger.Logger) ([]byte, error) {
	logger.Log("level", "info", "msg", fmt.Sprintf("Executing: %s %v", cmd, args))

	// Create cmd and add environment.
	c := exec.Command(cmd, args...)
	c.Env = append(os.Environ(), envs...)

	// Execute and get output.
	stdOutErr, err := c.CombinedOutput()
	if err != nil {
		logger.Log("level", "error", "msg", "execCmd failed", "reason", fmt.Sprintf("%s", stdOutErr), "err", err)
		return stdOutErr, microerror.Mask(err)
	}
	return stdOutErr, nil
}

// Encrypt data with passphrase.
func encryptData(value []byte, pass string) (ciphertext []byte, err error) {
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
func encryptFile(srcPath string, dstPart string, passphrase string) error {
	data, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return microerror.Mask(err)
	}

	encData, err := encryptData(data, passphrase)
	if err != nil {
		return microerror.Mask(err)
	}

	err = ioutil.WriteFile(dstPart, encData, os.FileMode(0600))
	if err != nil {
		return microerror.Mask(err)
	}

	return nil
}
