package etcd

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/giantswarm/microerror"
	"github.com/giantswarm/micrologger"
	"github.com/mholt/archiver/v3"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"github.com/giantswarm/etcd-backup-operator/v4/pkg/etcd/internal/encrypt"
	"github.com/giantswarm/etcd-backup-operator/v4/pkg/etcd/key"
	"github.com/giantswarm/etcd-backup-operator/v4/pkg/etcd/proxy"
)

type V3Backup struct {
	EncPass   string
	Endpoints string
	Logger    micrologger.Logger
	Prefix    string

	etcdClient *clientv3.Client
	filename   *string
	tmpDir     *string
}

func NewV3Backup(tlsConfig *tls.Config, p *proxy.Proxy, encPass string, endpoints string, logger micrologger.Logger, prefix string) (V3Backup, error) {
	filename := ""
	tmpDir := ""

	etcdClient, err := createEtcdV3Client(endpoints, tlsConfig, p)
	if err != nil {
		return V3Backup{}, microerror.Mask(err)
	}

	return V3Backup{
		EncPass:   encPass,
		Endpoints: endpoints,
		Logger:    logger,
		Prefix:    prefix,

		etcdClient: etcdClient,
		filename:   &filename,
		tmpDir:     &tmpDir,
	}, nil
}

func createEtcdV3Client(endpoint string, tlsConfig *tls.Config, p *proxy.Proxy) (*clientv3.Client, error) {
	dialOpt := []grpc.DialOption{
		grpc.WithBlock(), // block until the underlying connection is up
	}

	// add proxy dialer if proxy is not nil
	if p != nil {
		dialer, err := proxy.NewDialer(*p)
		if err != nil {
			return nil, microerror.Mask(err)
		}

		dialOpt = append(dialOpt, grpc.WithContextDialer(dialer.DialContextWithAddr))
	}

	c, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{endpoint},
		DialTimeout: time.Second * 60,
		DialOptions: dialOpt,
		TLS:         tlsConfig,
	})

	if err != nil {
		return nil, microerror.Mask(err)
	}

	return c, nil
}

// Cleanup clears temporary directory
func (b V3Backup) Cleanup() {
	os.RemoveAll(b.getTmpDir())
}

// Create etcd in temporary directory.
func (b V3Backup) Create() (string, error) {
	ctx := context.Background()
	err := b.compactAndDefrag()
	if err != nil {
		return "", microerror.Mask(err)
	}

	// filename
	*b.filename = b.Prefix + "-v3-" + time.Now().Format(key.TsFormat) + key.DbExt

	// Full path to file.
	fpath := filepath.Join(b.getTmpDir(), *b.filename)

	// Create a etcd.
	snapshot, err := b.etcdClient.Snapshot(ctx)
	if err != nil {
		return "", microerror.Mask(err)
	}

	outFile, err := os.Create(fpath)
	if err != nil {
		return "", microerror.Mask(err)
	}
	// handle err
	defer func() { _ = outFile.Close() }()
	_, err = io.Copy(outFile, snapshot)
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

	b.Logger.Log("level", "info", "msg", "Etcd v3 backup created successfully", "file", b.filename)
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
		tmpDir, err := os.MkdirTemp("", "")
		if err != nil {
			panic(err)
		}
		b.Logger.LogCtx(context.Background(), fmt.Sprintf("Created temporary directory: %s", tmpDir))
		*b.tmpDir = tmpDir
	}

	return *b.tmpDir
}

func (b V3Backup) compactAndDefrag() error {
	ctx := context.Background()
	b.Logger.Debugf(ctx, "Compacting etcd instance")
	// Get latest revision.
	s, err := b.etcdClient.Status(ctx, b.Endpoints)

	if err != nil {
		return microerror.Mask(err)
	}

	b.Logger.Debugf(context.Background(), "Revision is %d", s.Header.Revision)

	_, err = b.etcdClient.Compact(ctx, s.Header.Revision)

	if err != nil {
		return microerror.Mask(err)
	}

	b.Logger.Debugf(context.Background(), "Compacted etcd instance")

	b.Logger.Debugf(context.Background(), "Defragging etcd instance")

	_, err = b.etcdClient.Defragment(ctx, b.Endpoints)

	if err != nil {
		return microerror.Mask(err)
	}

	b.Logger.Debugf(context.Background(), "Defragged etcd instance")

	return nil
}
