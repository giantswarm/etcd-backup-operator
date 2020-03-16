#!/bin/sh

name="etcd-backup-$(date +'%Y%m%d%H%M%S')"
guestBackup="$1"

if [ "$guestBackup" != "true" ] && [ "$guestBackup" != "false" ]
then
  echo "Usage: ${0} <true|false>"
  exit 1
fi

TEMPLATE=$(cat <<-END
apiVersion: "backup.giantswarm.io/v1alpha1"
kind: "ETCDBackup"
metadata:
  name: "${name}"
spec:
  guestBackup: $guestBackup
END
)

echo "$TEMPLATE" | kubectl apply -f -
