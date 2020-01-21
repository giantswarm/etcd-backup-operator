package backup

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/mholt/archiver"
)

type EtcdBackupV3 struct {
	CACert    string
	Cert      string
	EncPass   string
	Endpoints string
	Filename  string
	Logger    micrologger.Logger
	Key       string
	Prefix    string
	TmpDir    string
}

func (b *EtcdBackupV3) getTmpDir() string {
	if len(b.TmpDir) == 0 {
		tmpDir, err := ioutil.TempDir("", "")
		if err != nil {
			panic(err)
		}
		b.Logger.LogCtx(context.Background(), fmt.Sprintf("Created temporary directory: %s", tmpDir))
		b.TmpDir = tmpDir
	}

	return b.TmpDir
}

//clear temporary directory
func (b *EtcdBackupV3) Cleanup() {
	os.RemoveAll(b.getTmpDir())
}

// Create etcd in temporary directory.
func (b *EtcdBackupV3) Create() (string, error) {
	// Filename
	b.Filename = b.Prefix + "-backup-etcd-v3-" + getTimeStamp() + dbExt

	// Full path to file.
	fpath := filepath.Join(b.getTmpDir(), b.Filename)

	etcdctlEnvs := []string{"ETCDCTL_API=3"}
	etcdctlArgs := []string{
		"snapshot",
		"save",
		fpath,
	}

	if b.Endpoints != "" {
		etcdctlArgs = append(etcdctlArgs, "--endpoints", b.Endpoints)
	}
	if b.CACert != "" {
		etcdctlArgs = append(etcdctlArgs, "--cacert", b.CACert)
	}
	if b.Cert != "" {
		etcdctlArgs = append(etcdctlArgs, "--cert", b.Cert)
	}
	if b.Key != "" {
		etcdctlArgs = append(etcdctlArgs, "--key", b.Key)
	}

	// Create a etcd.
	_, err := execCmd(etcdctlCmd, etcdctlArgs, etcdctlEnvs, b.Logger)
	if err != nil {
		return "", microerror.Mask(err)
	}

	// Create tar.gz.
	err = archiver.Archive([]string{fpath}, fpath+tgzExt)
	if err != nil {
		return "", microerror.Mask(err)
	}

	// Update Filename in etcd object.
	b.Filename = b.Filename + tgzExt
	fpath = filepath.Join(b.getTmpDir(), b.Filename)

	b.Logger.Log("level", "info", "msg", "Etcd v3 backup created successfully")
	return fpath, nil
}

// encrypt backup
func (b *EtcdBackupV3) Encrypt() (string, error) {
	// Full path to file.
	fpath := filepath.Join(b.getTmpDir(), b.Filename)

	if b.EncPass == "" {
		b.Logger.Log("level", "warning", "msg", "No passphrase provided. Skipping etcd v3 backup encryption")
		return fpath, nil
	}

	// Encrypt etcd.
	err := encryptFile(fpath, fpath+encExt, b.EncPass)
	if err != nil {
		return "", microerror.Mask(err)
	}

	// Update Filename in etcd object.
	b.Filename = b.Filename + encExt
	fpath = filepath.Join(b.getTmpDir(), b.Filename)

	b.Logger.Log("level", "info", "msg", "Etcd v3 backup encrypted successfully")
	return fpath, nil
}

func (b *EtcdBackupV3) Version() string {
	return "v3"
}
