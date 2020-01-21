package backup

import (
	"context"
	"fmt"
	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/mholt/archiver"
	"io/ioutil"
	"os"
	"path/filepath"
)

type EtcdBackupV2 struct {
	Datadir  string
	EncPass  string
	Filename string
	Logger   micrologger.Logger
	Prefix   string
	TmpDir   string
}

func (b *EtcdBackupV2) getTmpDir() string {
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
func (b *EtcdBackupV2) Cleanup() {
	os.RemoveAll(b.getTmpDir())
}

// Create etcd in temporary directory, tar and compress.
func (b *EtcdBackupV2) Create() (string, error) {
	// Filename
	b.Filename = b.Prefix + "-etcd-etcd-v2-" + getTimeStamp()

	// Full path to file.
	fpath := filepath.Join(b.getTmpDir(), b.Filename)

	// Create a etcd.
	etcdctlEnvs := []string{}
	etcdctlArgs := []string{
		"backup",
		"--data-dir", b.Datadir,
		"--backup-dir", fpath,
	}

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

	b.Logger.Log("level", "info", "msg", "Etcd v2 etcd created successfully")
	return fpath, nil
}

func (b *EtcdBackupV2) Encrypt() (string, error) {
	// Full path to file.
	fpath := filepath.Join(b.getTmpDir(), b.Filename)

	if b.EncPass == "" {
		b.Logger.Log("level", "warning", "msg", "No passphrase provided. Skipping etcd v2 backup encryption")
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

	b.Logger.Log("level", "info", "msg", "Etcd v2 backup encrypted successfully")
	return fpath, nil
}

func (b *EtcdBackupV2) Version() string {
	return "v2"
}
