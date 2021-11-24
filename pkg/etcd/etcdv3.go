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
	"k8s.io/apimachinery/pkg/util/json"

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
	err := b.compactAndDefrag()
	if err != nil {
		return "", microerror.Mask(err)
	}

	// filename
	*b.filename = b.Prefix + "-v3-" + time.Now().Format(key.TsFormat) + key.DbExt

	// Full path to file.
	fpath := filepath.Join(b.getTmpDir(), *b.filename)

	etcdctlArgs := []string{
		"snapshot",
		"save",
		fpath,
		"--dial-timeout=10s",
		"--command-timeout=30s",
	}

	// Create a etcd.
	_, err = b.runEtcdctlCmd(etcdctlArgs)
	if err != nil {
		return "", microerror.Mask(err)
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

func (b V3Backup) compactAndDefrag() error {
	b.Logger.Debugf(context.Background(), "Compacting etcd instance")
	// Get latest revision.
	output, err := b.runEtcdctlCmd([]string{
		"endpoint",
		"status",
		`--write-out=json`,
	})
	if err != nil {
		return microerror.Mask(err)
	}

	revision, err := getRevision(output)
	if err != nil {
		return microerror.Mask(err)
	}

	b.Logger.Debugf(context.Background(), "Revision is %d", revision)

	_, err = b.runEtcdctlCmd([]string{
		"compact",
		fmt.Sprintf("%d", revision),
	})
	if err != nil {
		return microerror.Mask(err)
	}

	b.Logger.Debugf(context.Background(), "Compacted etcd instance")

	b.Logger.Debugf(context.Background(), "Defragging etcd instance")

	_, err = b.runEtcdctlCmd([]string{
		"defrag",
		"--command-timeout=60s",
		"--dial-timeout=60s",
		"--keepalive-timeout=25s",
	})
	if err != nil {
		return microerror.Mask(err)
	}

	b.Logger.Debugf(context.Background(), "Defragged etcd instance")

	return nil
}

func (b V3Backup) runEtcdctlCmd(etcdctlArgs []string) ([]byte, error) {
	etcdctlEnvs := []string{"ETCDCTL_API=3"}

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

	log, err := exec.Cmd(key.EtcdctlCmd, etcdctlArgs, etcdctlEnvs, b.Logger)
	if err != nil {
		return nil, errors.New(string(log))
	}

	return log, nil
}

func getRevision(output []byte) (int32, error) {
	type endpointStatusOutputStatusHeader struct {
		Revision int32
	}

	type endpointStatusOutputStatus struct {
		Header endpointStatusOutputStatusHeader
	}

	type endpointStatusOutput struct {
		Status endpointStatusOutputStatus
	}

	// [
	//  {
	//    "Endpoint": "https://127.0.0.1:2379",
	//    "Status": {
	//      "header": {
	//        "cluster_id": 14841639068965180000,
	//        "member_id": 11895648879011906000,
	//        "revision": 197531277,
	//        "raft_term": 48069
	//      },
	//      "version": "3.4.13",
	//      "dbSize": 79740928,
	//      "leader": 1412153380952468500,
	//      "raftIndex": 224800332,
	//      "raftTerm": 48069
	//    }
	//  }
	//]

	var status []endpointStatusOutput
	err := json.Unmarshal(output, &status)
	if err != nil {
		return 0, microerror.Mask(err)
	}

	if len(status) == 0 {
		return 0, microerror.Maskf(emptyEndpointHealthError, "The etcdctl endpoint status command returned zero results.")
	}

	return status[0].Status.Header.Revision, nil
}
