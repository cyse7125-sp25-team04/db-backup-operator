/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BackupDatabaseSchemaSpec defines the desired state of BackupDatabaseSchema
type BackupDatabaseSchemaSpec struct {
    DBHost                     string `json:"dbHost"`
    DBUser                     string `json:"dbUser"`
    DBPasswordSecretName       string `json:"dbPasswordSecretName"`
    DBPasswordSecretNamespace  string `json:"dbPasswordSecretNamespace"`
    DBPasswordSecretKey        string `json:"dbPasswordSecretKey"`
    DBName                     string `json:"dbName"`
    DBSchema                   string `json:"dbSchema"`
    DBPort                     int32  `json:"dbPort"`
    GCSBucket                  string `json:"gcsBucket"`
    KubeServiceAccount         string `json:"kubeServiceAccount"`
    GCPServiceAccount          string `json:"gcpServiceAccount"`
    BackupJobNamespace         string `json:"backupJobNamespace"`
}

// BackupDatabaseSchemaStatus defines the observed state of BackupDatabaseSchema
type BackupDatabaseSchemaStatus struct {
    LastBackupTime string `json:"lastBackupTime,omitempty"`  // UTC time stamp of the last backup run
    BackupLocation string `json:"backupLocation,omitempty"`  // Full path in the GCS bucket
    JobStatus      string `json:"jobStatus,omitempty"`       // e.g., success, failed, running
    RecentJobName  string `json:"recentJobName,omitempty"`
}


// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// BackupDatabaseSchema is the Schema for the backupdatabaseschemas API.
type BackupDatabaseSchema struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackupDatabaseSchemaSpec   `json:"spec,omitempty"`
	Status BackupDatabaseSchemaStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BackupDatabaseSchemaList contains a list of BackupDatabaseSchema.
type BackupDatabaseSchemaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackupDatabaseSchema `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackupDatabaseSchema{}, &BackupDatabaseSchemaList{})
}
