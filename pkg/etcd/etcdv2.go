package etcd

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/mholt/archiver"

	"github.com/giantswarm/etcd-backup-operator/pkg/etcd/internal/encrypt"
	"github.com/giantswarm/etcd-backup-operator/pkg/etcd/internal/exec"
	"github.com/giantswarm/etcd-backup-operator/pkg/etcd/key"
)

type V2Backup struct {
	Datadir  string
	EncPass  string
	Filename string
	Logger   micrologger.Logger
	Prefix   string
	TmpDir   string
}

// Clear temporary directory.
func (b V2Backup) Cleanup() {
	os.RemoveAll(b.getTmpDir())
}

// Create etcd in temporary directory, tar and compress.
func (b V2Backup) Create() (string, error) {
	// Filename.
	b.Filename = b.Prefix + "-etcd-etcd-v2-" + time.Now().Format(key.TsFormat)

	// Full path to file.
	fpath := filepath.Join(b.getTmpDir(), b.Filename)

	// Create a etcd.
	etcdctlEnvs := []string{}
	etcdctlArgs := []string{
		"backup",
		"--data-dir", b.Datadir,
		"--backup-dir", fpath,
	}

	log, err := exec.Cmd(key.EtcdctlCmd, etcdctlArgs, etcdctlEnvs, b.Logger)
	if err != nil {
		return "", errors.New(string(log))
	}

	// Create tar.gz.
	err = archiver.Archive([]string{fpath}, fpath+key.TgzExt)
	if err != nil {
		return "", microerror.Mask(err)
	}

	// Update Filename in etcd object.
	b.Filename = b.Filename + key.TgzExt
	fpath = filepath.Join(b.getTmpDir(), b.Filename)

	b.Logger.Log("level", "info", "msg", "Etcd v2 etcd created successfully")
	return fpath, nil
}

// Encrypts the backup file.
func (b V2Backup) Encrypt() (string, error) {
	// Full path to file.
	fpath := filepath.Join(b.getTmpDir(), b.Filename)

	if b.EncPass == "" {
		b.Logger.Log("level", "warning", "msg", "No passphrase provided. Skipping etcd v2 backup encryption")
		return fpath, nil
	}

	// Encrypt etcd.
	err := encrypt.File(fpath, fpath+key.EncExt, b.EncPass)
	if err != nil {
		return "", microerror.Mask(err)
	}

	// Update Filename in etcd object.
	b.Filename = b.Filename + key.EncExt
	fpath = filepath.Join(b.getTmpDir(), b.Filename)

	b.Logger.Log("level", "info", "msg", "Etcd v2 backup encrypted successfully")
	return fpath, nil
}

func (b V2Backup) Version() string {
	return "v2"
}

func (b V2Backup) getTmpDir() string {
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
