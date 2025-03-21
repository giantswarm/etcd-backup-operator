#!/usr/bin/env sh
# Get name, guest backup & clusters regex.
name="etcd-backup-$(date "+%Y%m%d%H%M%S")"
guest_backup="${1}"
clusters_regex="${2:-.*}"
clusters_to_exclude_regex="${3:-^$}"
destination="${4:-primary}" # Add destination parameter, default to "primary"

# Check guest backup.
if [ "${guest_backup}" != "true" ] && [ "${guest_backup}" != "false" ]
then
  # Print usage.
  echo "Usage: ${0} <true|false> [clusters_regex] [clusters_to_exclude_regex] [destination]"
  # Exit erroneously.
  exit 1
fi

# Create etcd backup.
kubectl create --filename - <<END
apiVersion: backup.giantswarm.io/v1alpha1
kind: ETCDBackup
metadata:
  name: ${name}
  labels:
    backup.giantswarm.io/destination: ${destination}
spec:
  guestBackup: ${guest_backup}
  clustersRegex: "${clusters_regex}"
  clustersToExcludeRegex: "${clusters_to_exclude_regex}"
END
