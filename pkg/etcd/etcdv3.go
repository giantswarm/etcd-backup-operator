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
	"github.com/mholt/archiver/v3"

	"github.com/giantswarm/etcd-backup-operator/v2/pkg/etcd/internal/encrypt"
	"github.com/giantswarm/etcd-backup-operator/v2/pkg/etcd/internal/exec"
	"github.com/giantswarm/etcd-backup-operator/v2/pkg/etcd/key"
)

type V3Backup struct {
	CACert    string
	Cert      string
	EncPass   string
	Endpoints string
	Logger    micrologger.Logger
	Key       string
	Prefix    string

	filename *string
	tmpDir   *string
}

func NewV3Backup(caCert string, cert string, encPass string, endpoints string, logger micrologger.Logger, key string, prefix string) V3Backup {
	filename := ""
	tmpDir := ""

	return V3Backup{
		CACert:    caCert,
		Cert:      cert,
		EncPass:   encPass,
		Endpoints: endpoints,
		Logger:    logger,
		Key:       key,
		Prefix:    prefix,

		filename: &filename,
		tmpDir:   &tmpDir,
	}
}

//clear temporary directory
func (b V3Backup) Cleanup() {
	os.RemoveAll(b.getTmpDir())
}

// Create etcd in temporary directory.
func (b V3Backup) Create() (string, error) {
	// filename
	*b.filename = b.Prefix + "-backup-etcd-v3-" + time.Now().Format(key.TsFormat) + key.DbExt

	// Full path to file.
	fpath := filepath.Join(b.getTmpDir(), *b.filename)

	etcdctlEnvs := []string{"ETCDCTL_API=3"}
	etcdctlArgs := []string{
		"snapshot",
		"save",
		fpath,
		"--dial-timeout=10s",
		"--command-timeout=30s",
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
	log, err := exec.Cmd(key.EtcdctlCmd, etcdctlArgs, etcdctlEnvs, b.Logger)
	if err != nil {
		return "", errors.New(string(log))
	}

	// Create tar.gz.
	err = archiver.Archive([]string{fpath}, fpath+key.TgzExt)
	if err != nil {
		return "", microerror.Mask(err)
	}

	// Update filename in etcd object.
	*b.filename = *b.filename + key.TgzExt
	fpath = filepath.Join(b.getTmpDir(), *b.filename)

	b.Logger.Log("level", "info", "msg", "Etcd v3 backup created successfully")
	return fpath, nil
}

// Encrypt backup.
func (b V3Backup) Encrypt() (string, error) {
	// Full path to file.
	fpath := filepath.Join(b.getTmpDir(), *b.filename)

	if b.EncPass == "" {
		b.Logger.Log("level", "warning", "msg", "No passphrase provided. Skipping etcd v3 backup encryption")
		return fpath, nil
	}

	// Encrypt etcd.
	err := encrypt.File(fpath, fpath+key.EncExt, b.EncPass)
	if err != nil {
		return "", microerror.Mask(err)
	}

	// Update filename in etcd object.
	*b.filename = *b.filename + key.EncExt
	fpath = filepath.Join(b.getTmpDir(), *b.filename)

	b.Logger.Log("level", "info", "msg", "Etcd v3 backup encrypted successfully")
	return fpath, nil
}

func (b V3Backup) Version() string {
	return "v3"
}

func (b V3Backup) getTmpDir() string {
	if len(*b.tmpDir) == 0 {
		tmpDir, err := ioutil.TempDir("", "")
		if err != nil {
			panic(err)
		}
		b.Logger.LogCtx(context.Background(), fmt.Sprintf("Created temporary directory: %s", tmpDir))
		*b.tmpDir = tmpDir
	}

	return *b.tmpDir
}
