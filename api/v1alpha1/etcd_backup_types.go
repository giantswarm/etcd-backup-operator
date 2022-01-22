package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=common;giantswarm,scope=Cluster
// +k8s:openapi-gen=true

type ETCDBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	Spec              ETCDBackupSpec `json:"spec"`
	// +kubebuilder:validation:Optional
	Status ETCDBackupStatus `json:"status,omitempty"`
}

// +k8s:openapi-gen=true
type ETCDBackupSpec struct {
	// +kubebuilder:validation:Optional
	// +nullable
	// GuestBackup is a boolean indicating if the workload clusters have to be backed up
	GuestBackup bool `json:"guestBackup"`

	// +kubebuilder:validation:Optional
	// +nullable
	// ClusterNames is a list of cluster IDs that should be backed up. Can contain the special value 'ManagementCluster' to indicate the Management Cluster.
	ClusterNames []string `json:"clusterNames,omitempty"`
}

// +k8s:openapi-gen=true
type ETCDBackupStatus struct {
	// +kubebuilder:validation:Optional
	// +nullable
	// map containing the state of the backup for all instances
	Instances map[string]ETCDInstanceBackupStatusIndex `json:"instances,omitempty"`
	// Status of the whole backup job (can be 'Pending', 'Running'. 'Completed', 'Failed')
	Status string `json:"status"`
	// +kubebuilder:validation:Optional
	// +nullable
	// Timestamp when the first attempt was made
	StartedTimestamp metav1.Time `json:"startedTimestamp,omitempty"`
	// +kubebuilder:validation:Optional
	// +nullable
	// Timestamp when the last (final) attempt was made (when the Phase became either 'Completed' or 'Failed'
	FinishedTimestamp metav1.Time `json:"finishedTimestamp,omitempty"`
}

// +k8s:openapi-gen=true
type ETCDInstanceBackupStatusIndex struct {
	// Name of the workload cluster or management cluster
	Name string `json:"name"`
	// +kubebuilder:validation:Optional
	// +nullable
	// Error details in case the backup is failed.
	Error string `json:"error,omitempty"`
	// +kubebuilder:validation:Optional
	// +nullable
	// Status of the V2 backup for this instance
	V2 *ETCDInstanceBackupStatus `json:"v2,omitempty"`
	// +kubebuilder:validation:Optional
	// +nullable
	// Status of the V3 backup for this instance
	V3 *ETCDInstanceBackupStatus `json:"v3,omitempty"`
}

// +k8s:openapi-gen=true
type ETCDInstanceBackupStatus struct {
	// Status of this instance's backup job (can be 'Pending', 'Running'. 'Completed', 'Failed')
	Status string `json:"status"`
	// +kubebuilder:validation:Optional
	// +nullable
	// Timestamp when the first attempt was made
	StartedTimestamp metav1.Time `json:"startedTimestamp,omitempty"`
	// +kubebuilder:validation:Optional
	// +nullable
	// Timestamp when the last (final) attempt was made (when the Phase became either 'Completed' or 'Failed'
	FinishedTimestamp metav1.Time `json:"finishedTimestamp,omitempty"`
	// +kubebuilder:validation:Optional
	// Latest backup error message
	LatestError string `json:"latestError,omitempty"`
	// +kubebuilder:validation:Optional
	// Time took by the backup creation process
	CreationTime int64 `json:"creationTime,omitempty"`
	// +kubebuilder:validation:Optional
	// Time took by the backup encryption process
	EncryptionTime int64 `json:"encryptionTime,omitempty"`
	// +kubebuilder:validation:Optional
	// Time took by the backup upload process
	UploadTime int64 `json:"uploadTime,omitempty"`
	// +kubebuilder:validation:Optional
	// Size of the backup file
	BackupFileSize int64 `json:"backupFileSize,omitempty"`
	// +kubebuilder:validation:Optional
	// +nullable
	// Filename is the name of the backup file.
	Filename string `json:"filename,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ETCDBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []ETCDBackup `json:"items"`
}
