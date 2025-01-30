#!/usr/bin/env sh
# Get name, guest backup & clusters regex.
name="etcd-backup-$(date "+%Y%m%d%H%M%S")"
guest_backup="${1}"
clusters_regex="${2:-.*}"
clusters_to_exclude_regex="${3:-.*}"

# Check guest backup.
if [ "${guest_backup}" != "true" ] && [ "${guest_backup}" != "false" ]
then
  # Print usage.
  echo "Usage: ${0} <true|false> [clusters_regex] [clusters_to_exclude_regex]"
  # Exit erroneously.
  exit 1
fi

# Create etcd backup.
kubectl create --filename - <<END
apiVersion: backup.giantswarm.io/v1alpha1
kind: ETCDBackup
metadata:
  name: ${name}
spec:
  guestBackup: ${guest_backup}
  clustersRegex: "${clusters_regex}"
  clustersToExcludeRegex: "${clusters_to_exclude_regex}"
END
