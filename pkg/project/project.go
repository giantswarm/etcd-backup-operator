package project

var (
	description = "The etcd-backup-operator does something."
	gitSHA      = "n/a"
	name        = "etcd-backup-operator"
	source      = "https://github.com/giantswarm/etcd-backup-operator"
	version     = "4.4.3-dev"
)

func Description() string {
	return description
}

func GitSHA() string {
	return gitSHA
}

func Name() string {
	return name
}

func Source() string {
	return source
}

func Version() string {
	return version
}
