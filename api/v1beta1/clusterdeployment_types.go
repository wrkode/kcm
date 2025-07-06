// Copyright 2024
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1beta1

import (
	"encoding/json"
	"fmt"

	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

const (
	BlockingFinalizer          = "k0rdent.mirantis.com/cleanup"
	ClusterDeploymentFinalizer = "k0rdent.mirantis.com/cluster-deployment"

	FluxHelmChartNameKey      = "helm.toolkit.fluxcd.io/name"
	FluxHelmChartNamespaceKey = "helm.toolkit.fluxcd.io/namespace"

	KCMManagedLabelKey   = "k0rdent.mirantis.com/managed"
	KCMManagedLabelValue = "true"
)

const (
	// ClusterDeploymentKind is the string representation of a ClusterDeployment.
	ClusterDeploymentKind = "ClusterDeployment"
	// TemplateReadyCondition indicates the referenced Template exists and valid.
	TemplateReadyCondition = "TemplateReady"
	// HelmChartReadyCondition indicates the corresponding HelmChart is valid and ready.
	HelmChartReadyCondition = "HelmChartReady"
	// HelmReleaseReadyCondition indicates the corresponding HelmRelease is ready and fully reconciled.
	HelmReleaseReadyCondition = "HelmReleaseReady"
	// CAPIClusterSummaryCondition aggregates all the important conditions from the Cluster object.
	CAPIClusterSummaryCondition = "CAPIClusterSummary"
	// SveltosClusterReadyCondition indicates the sveltos cluster is valid and ready.
	SveltosClusterReadyCondition = "SveltosClusterReady"
)

// ClusterDeploymentSpec defines the desired state of ClusterDeployment
type ClusterDeploymentSpec struct {
	// Config allows to provide parameters for template customization.
	// If no Config provided, the field will be populated with the default values for
	// the template and DryRun will be enabled.
	Config *apiextv1.JSON `json:"config,omitempty"`

	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253

	// Template is a reference to a Template object located in the same namespace.
	Template string `json:"template"`
	// Name reference to the related Credentials object.
	Credential string `json:"credential,omitempty"`
	// IPAMClaim defines IP Address Management (IPAM) requirements for the cluster.
	// It can either reference an existing IPAM claim or specify an inline claim.
	IPAMClaim ClusterIPAMClaimType `json:"ipamClaim,omitempty"`
	// ServiceSpec is spec related to deployment of services.
	ServiceSpec ServiceSpec `json:"serviceSpec,omitempty"`
	// DryRun specifies whether the template should be applied after validation or only validated.
	DryRun bool `json:"dryRun,omitempty"`
	// +kubebuilder:default:=true

	// PropagateCredentials indicates whether credentials should be propagated
	// for use by CCM (Cloud Controller Manager).
	PropagateCredentials bool `json:"propagateCredentials,omitempty"`

	// AdoptedCluster defines configuration for adopted cluster lifecycle management.
	// This enables full infrastructure lifecycle management for adopted clusters.
	AdoptedCluster *AdoptedClusterConfig `json:"adoptedCluster,omitempty"`
}

// AdoptedClusterConfig defines the configuration for adopted cluster lifecycle management.
type AdoptedClusterConfig struct {
	// Enabled indicates whether adopted cluster lifecycle management is enabled.
	// +kubebuilder:default:=false
	Enabled bool `json:"enabled"`

	// DiscoveryMode defines how infrastructure should be discovered.
	// +kubebuilder:validation:Enum=auto;manual
	// +kubebuilder:default:=auto
	DiscoveryMode string `json:"discoveryMode,omitempty"`

	// InfrastructureMapping defines the mapping of existing infrastructure.
	InfrastructureMapping *InfrastructureMapping `json:"infrastructureMapping,omitempty"`

	// LifecycleManagement defines lifecycle management capabilities.
	LifecycleManagement *LifecycleManagement `json:"lifecycleManagement,omitempty"`
}

// InfrastructureMapping defines the mapping of existing infrastructure to CAPI resources.
type InfrastructureMapping struct {
	// Provider specifies the cloud provider for the adopted cluster.
	// +kubebuilder:validation:Enum=aws;azure;gcp;vsphere;openstack;unknown
	Provider string `json:"provider"`

	// Region specifies the cloud region where the cluster is located.
	Region string `json:"region,omitempty"`

	// Credentials specifies the credentials for cloud provider access.
	Credentials *CloudCredentials `json:"credentials,omitempty"`

	// NodeMapping defines how existing nodes should be mapped to CAPI Machine resources.
	NodeMapping []NodeMapping `json:"nodeMapping,omitempty"`
}

// CloudCredentials defines credentials for cloud provider access.
type CloudCredentials struct {
	// SecretName specifies the name of the secret containing cloud credentials.
	SecretName string `json:"secretName"`

	// Namespace specifies the namespace where the secret is located.
	// +kubebuilder:default:=default
	Namespace string `json:"namespace,omitempty"`
}

// NodeMapping defines how an existing node should be mapped to a CAPI Machine resource.
type NodeMapping struct {
	// NodeName specifies the name of the existing node.
	NodeName string `json:"nodeName"`

	// MachineName specifies the name for the generated CAPI Machine resource.
	MachineName string `json:"machineName"`

	// InfrastructureRef specifies the infrastructure reference for the machine.
	InfrastructureRef *InfrastructureReference `json:"infrastructureRef,omitempty"`
}

// InfrastructureReference defines a reference to an infrastructure resource.
type InfrastructureReference struct {
	// APIVersion specifies the API version of the infrastructure resource.
	APIVersion string `json:"apiVersion"`

	// Kind specifies the kind of the infrastructure resource.
	Kind string `json:"kind"`

	// Name specifies the name of the infrastructure resource.
	Name string `json:"name"`
}

// LifecycleManagement defines lifecycle management capabilities for adopted clusters.
type LifecycleManagement struct {
	// Enabled indicates whether lifecycle management is enabled.
	// +kubebuilder:default:=false
	Enabled bool `json:"enabled"`

	// Scaling defines scaling capabilities for the adopted cluster.
	Scaling *ScalingConfig `json:"scaling,omitempty"`

	// Upgrades defines upgrade capabilities for the adopted cluster.
	Upgrades *UpgradeConfig `json:"upgrades,omitempty"`

	// Monitoring defines monitoring capabilities for the adopted cluster.
	Monitoring *MonitoringConfig `json:"monitoring,omitempty"`

	// Backup defines backup capabilities for the adopted cluster.
	Backup *BackupConfig `json:"backup,omitempty"`
}

// ScalingConfig defines scaling capabilities for adopted clusters.
type ScalingConfig struct {
	// Enabled indicates whether scaling is enabled.
	// +kubebuilder:default:=false
	Enabled bool `json:"enabled"`

	// AutoScaling indicates whether auto-scaling is enabled.
	// +kubebuilder:default:=false
	AutoScaling bool `json:"autoScaling"`

	// MinNodes specifies the minimum number of nodes.
	// +kubebuilder:validation:Minimum=1
	MinNodes int32 `json:"minNodes,omitempty"`

	// MaxNodes specifies the maximum number of nodes.
	// +kubebuilder:validation:Minimum=1
	MaxNodes int32 `json:"maxNodes,omitempty"`

	// NodeGroups defines different node groups for scaling.
	NodeGroups []NodeGroup `json:"nodeGroups,omitempty"`
}

// NodeGroup defines a group of nodes with similar characteristics.
type NodeGroup struct {
	// Name specifies the name of the node group.
	Name string `json:"name"`

	// InstanceType specifies the instance type for new nodes.
	InstanceType string `json:"instanceType,omitempty"`

	// MinNodes specifies the minimum number of nodes in this group.
	// +kubebuilder:validation:Minimum=0
	MinNodes int32 `json:"minNodes,omitempty"`

	// MaxNodes specifies the maximum number of nodes in this group.
	// +kubebuilder:validation:Minimum=1
	MaxNodes int32 `json:"maxNodes,omitempty"`

	// Labels specifies labels to apply to nodes in this group.
	Labels map[string]string `json:"labels,omitempty"`

	// Taints specifies taints to apply to nodes in this group.
	Taints []Taint `json:"taints,omitempty"`
}

// Taint defines a taint to apply to nodes.
type Taint struct {
	// Key specifies the taint key.
	Key string `json:"key"`

	// Value specifies the taint value.
	Value string `json:"value,omitempty"`

	// Effect specifies the taint effect.
	// +kubebuilder:validation:Enum=NoSchedule;PreferNoSchedule;NoExecute
	Effect string `json:"effect"`
}

// UpgradeConfig defines upgrade capabilities for adopted clusters.
type UpgradeConfig struct {
	// Enabled indicates whether upgrades are enabled.
	// +kubebuilder:default:=false
	Enabled bool `json:"enabled"`

	// AutoUpgrade indicates whether automatic upgrades are enabled.
	// +kubebuilder:default:=false
	AutoUpgrade bool `json:"autoUpgrade"`

	// MaintenanceWindow specifies the maintenance window for upgrades.
	MaintenanceWindow string `json:"maintenanceWindow,omitempty"`

	// UpgradeStrategy specifies the upgrade strategy to use.
	// +kubebuilder:validation:Enum=rolling;in-place
	// +kubebuilder:default:=rolling
	UpgradeStrategy string `json:"upgradeStrategy,omitempty"`

	// MaxUnavailable specifies the maximum number of unavailable nodes during upgrade.
	// +kubebuilder:validation:Minimum=1
	MaxUnavailable int32 `json:"maxUnavailable,omitempty"`
}

// MonitoringConfig defines monitoring capabilities for adopted clusters.
type MonitoringConfig struct {
	// Enabled indicates whether monitoring is enabled.
	// +kubebuilder:default:=false
	Enabled bool `json:"enabled"`

	// Metrics defines which metrics to collect.
	Metrics []string `json:"metrics,omitempty"`

	// HealthChecks defines health check configurations.
	HealthChecks *HealthCheckConfig `json:"healthChecks,omitempty"`
}

// HealthCheckConfig defines health check configurations.
type HealthCheckConfig struct {
	// Enabled indicates whether health checks are enabled.
	// +kubebuilder:default:=true
	Enabled bool `json:"enabled"`

	// Interval specifies the health check interval.
	// +kubebuilder:default="30s"
	Interval string `json:"interval,omitempty"`

	// Timeout specifies the health check timeout.
	// +kubebuilder:default="10s"
	Timeout string `json:"timeout,omitempty"`

	// FailureThreshold specifies the number of failures before marking unhealthy.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default:=3
	FailureThreshold int32 `json:"failureThreshold,omitempty"`
}

// BackupConfig defines backup capabilities for adopted clusters.
type BackupConfig struct {
	// Enabled indicates whether backup is enabled.
	// +kubebuilder:default:=false
	Enabled bool `json:"enabled"`

	// Schedule specifies the backup schedule in cron format.
	Schedule string `json:"schedule,omitempty"`

	// Retention specifies how long to retain backups.
	Retention string `json:"retention,omitempty"`

	// Storage defines backup storage configuration.
	Storage *BackupStorage `json:"storage,omitempty"`
}

// BackupStorage defines backup storage configuration.
type BackupStorage struct {
	// Type specifies the storage type.
	// +kubebuilder:validation:Enum=s3;gcs;azure;local
	Type string `json:"type"`

	// Location specifies the storage location.
	Location string `json:"location"`

	// Credentials specifies credentials for storage access.
	Credentials *CloudCredentials `json:"credentials,omitempty"`
}

// ClusterIPAMClaimType represents the IPAM claim configuration for a cluster deployment.
// It allows referencing an existing claim or defining a new one inline.
type ClusterIPAMClaimType struct {
	// ClusterIPAMClaimSpec defines the inline IPAM claim specification if no reference is provided.
	// This allows for dynamic IP address allocation during cluster provisioning.
	ClusterIPAMClaimSpec *ClusterIPAMClaimSpec `json:"spec,omitempty"`

	// ClusterIPAMClaimRef is the name of an existing ClusterIPAMClaim resource to use.
	ClusterIPAMClaimRef string `json:"ref,omitempty"`
}

// ClusterDeploymentStatus defines the observed state of ClusterDeployment
type ClusterDeploymentStatus struct {
	// Services contains details for the state of services.
	Services []ServiceStatus `json:"services,omitempty"`
	// ServicesUpgradePaths contains details for the state of services upgrade paths.
	ServicesUpgradePaths []ServiceUpgradePaths `json:"servicesUpgradePaths,omitempty"`
	// Currently compatible exact Kubernetes version of the cluster. Being set only if
	// provided by the corresponding ClusterTemplate.
	KubernetesVersion string `json:"k8sVersion,omitempty"`
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type

	// Conditions contains details for the current state of the ClusterDeployment.
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// AvailableUpgrades is the list of ClusterTemplate names to which
	// this cluster can be upgraded. It can be an empty array, which means no upgrades are
	// available.
	AvailableUpgrades []string `json:"availableUpgrades,omitempty"`
	// ObservedGeneration is the last observed generation.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// AdoptedClusterStatus contains status information for adopted cluster lifecycle management.
	AdoptedClusterStatus *AdoptedClusterStatus `json:"adoptedClusterStatus,omitempty"`
}

// AdoptedClusterStatus defines the status of adopted cluster lifecycle management.
type AdoptedClusterStatus struct {
	// DiscoveryStatus contains status information for infrastructure discovery.
	DiscoveryStatus *DiscoveryStatus `json:"discoveryStatus,omitempty"`

	// CAPIStatus contains status information for CAPI integration.
	CAPIStatus *CAPIStatus `json:"capiStatus,omitempty"`

	// ScalingStatus contains status information for scaling operations.
	ScalingStatus *ScalingStatus `json:"scalingStatus,omitempty"`

	// UpgradeStatus contains status information for upgrade operations.
	UpgradeStatus *UpgradeStatus `json:"upgradeStatus,omitempty"`

	// MonitoringStatus contains status information for monitoring.
	MonitoringStatus *MonitoringStatus `json:"monitoringStatus,omitempty"`

	// BackupStatus contains status information for backup operations.
	BackupStatus *BackupStatus `json:"backupStatus,omitempty"`
}

// DiscoveryStatus defines the status of infrastructure discovery.
type DiscoveryStatus struct {
	// Phase indicates the current phase of discovery.
	// +kubebuilder:validation:Enum=NotStarted;InProgress;Completed;Failed
	Phase string `json:"phase"`

	// Message provides additional information about the discovery status.
	Message string `json:"message,omitempty"`

	// LastDiscoveryTime indicates when the last discovery was performed.
	LastDiscoveryTime *metav1.Time `json:"lastDiscoveryTime,omitempty"`

	// DiscoveredResources contains information about discovered infrastructure resources.
	DiscoveredResources []DiscoveredResource `json:"discoveredResources,omitempty"`
}

// DiscoveredResource defines a discovered infrastructure resource.
type DiscoveredResource struct {
	// Type specifies the type of discovered resource.
	Type string `json:"type"`

	// Name specifies the name of the discovered resource.
	Name string `json:"name"`

	// Provider specifies the cloud provider for this resource.
	Provider string `json:"provider,omitempty"`

	// Region specifies the region where this resource is located.
	Region string `json:"region,omitempty"`

	// Status indicates the status of this resource.
	Status string `json:"status,omitempty"`
}

// CAPIStatus defines the status of CAPI integration.
type CAPIStatus struct {
	// Phase indicates the current phase of CAPI integration.
	// +kubebuilder:validation:Enum=NotStarted;InProgress;Completed;Failed
	Phase string `json:"phase"`

	// Message provides additional information about the CAPI integration status.
	Message string `json:"message,omitempty"`

	// LastSyncTime indicates when the last CAPI sync was performed.
	LastSyncTime *metav1.Time `json:"lastSyncTime,omitempty"`

	// GeneratedResources contains information about generated CAPI resources.
	GeneratedResources []GeneratedResource `json:"generatedResources,omitempty"`
}

// GeneratedResource defines a generated CAPI resource.
type GeneratedResource struct {
	// APIVersion specifies the API version of the generated resource.
	APIVersion string `json:"apiVersion"`

	// Kind specifies the kind of the generated resource.
	Kind string `json:"kind"`

	// Name specifies the name of the generated resource.
	Name string `json:"name"`

	// Namespace specifies the namespace of the generated resource.
	Namespace string `json:"namespace,omitempty"`

	// Status indicates the status of this resource.
	Status string `json:"status,omitempty"`
}

// ScalingStatus defines the status of scaling operations.
type ScalingStatus struct {
	// Phase indicates the current phase of scaling.
	// +kubebuilder:validation:Enum=NotStarted;InProgress;Completed;Failed
	Phase string `json:"phase"`

	// Message provides additional information about the scaling status.
	Message string `json:"message,omitempty"`

	// LastScalingTime indicates when the last scaling operation was performed.
	LastScalingTime *metav1.Time `json:"lastScalingTime,omitempty"`

	// CurrentNodes indicates the current number of nodes.
	CurrentNodes int32 `json:"currentNodes,omitempty"`

	// DesiredNodes indicates the desired number of nodes.
	DesiredNodes int32 `json:"desiredNodes,omitempty"`

	// ScalingOperations contains information about recent scaling operations.
	ScalingOperations []ScalingOperation `json:"scalingOperations,omitempty"`
}

// ScalingOperation defines a scaling operation.
type ScalingOperation struct {
	// Type specifies the type of scaling operation.
	// +kubebuilder:validation:Enum=scale-up;scale-down
	Type string `json:"type"`

	// Timestamp indicates when the scaling operation was performed.
	Timestamp metav1.Time `json:"timestamp"`

	// FromNodes indicates the number of nodes before scaling.
	FromNodes int32 `json:"fromNodes"`

	// ToNodes indicates the number of nodes after scaling.
	ToNodes int32 `json:"toNodes"`

	// Status indicates the status of this scaling operation.
	Status string `json:"status"`

	// Message provides additional information about the scaling operation.
	Message string `json:"message,omitempty"`
}

// UpgradeStatus defines the status of upgrade operations.
type UpgradeStatus struct {
	// Phase indicates the current phase of upgrade.
	// +kubebuilder:validation:Enum=NotStarted;InProgress;Completed;Failed
	Phase string `json:"phase"`

	// Message provides additional information about the upgrade status.
	Message string `json:"message,omitempty"`

	// LastUpgradeTime indicates when the last upgrade was performed.
	LastUpgradeTime *metav1.Time `json:"lastUpgradeTime,omitempty"`

	// CurrentVersion indicates the current Kubernetes version.
	CurrentVersion string `json:"currentVersion,omitempty"`

	// TargetVersion indicates the target Kubernetes version for upgrade.
	TargetVersion string `json:"targetVersion,omitempty"`

	// UpgradeOperations contains information about recent upgrade operations.
	UpgradeOperations []UpgradeOperation `json:"upgradeOperations,omitempty"`
}

// UpgradeOperation defines an upgrade operation.
type UpgradeOperation struct {
	// Type specifies the type of upgrade operation.
	// +kubebuilder:validation:Enum=control-plane;node-pool;infrastructure
	Type string `json:"type"`

	// Timestamp indicates when the upgrade operation was performed.
	Timestamp metav1.Time `json:"timestamp"`

	// FromVersion indicates the version before upgrade.
	FromVersion string `json:"fromVersion"`

	// ToVersion indicates the version after upgrade.
	ToVersion string `json:"toVersion"`

	// Status indicates the status of this upgrade operation.
	Status string `json:"status"`

	// Message provides additional information about the upgrade operation.
	Message string `json:"message,omitempty"`
}

// MonitoringStatus defines the status of monitoring.
type MonitoringStatus struct {
	// Phase indicates the current phase of monitoring.
	// +kubebuilder:validation:Enum=NotStarted;InProgress;Completed;Failed
	Phase string `json:"phase"`

	// Message provides additional information about the monitoring status.
	Message string `json:"message,omitempty"`

	// LastHealthCheckTime indicates when the last health check was performed.
	LastHealthCheckTime *metav1.Time `json:"lastHealthCheckTime,omitempty"`

	// HealthStatus indicates the overall health status of the cluster.
	HealthStatus string `json:"healthStatus,omitempty"`

	// Metrics contains collected metrics information.
	Metrics []MetricInfo `json:"metrics,omitempty"`
}

// MetricInfo defines metric information.
type MetricInfo struct {
	// Name specifies the name of the metric.
	Name string `json:"name"`

	// Value specifies the current value of the metric.
	Value string `json:"value,omitempty"`

	// Unit specifies the unit of the metric.
	Unit string `json:"unit,omitempty"`

	// Timestamp indicates when this metric was collected.
	Timestamp metav1.Time `json:"timestamp"`
}

// BackupStatus defines the status of backup operations.
type BackupStatus struct {
	// Phase indicates the current phase of backup.
	// +kubebuilder:validation:Enum=NotStarted;InProgress;Completed;Failed
	Phase string `json:"phase"`

	// Message provides additional information about the backup status.
	Message string `json:"message,omitempty"`

	// LastBackupTime indicates when the last backup was performed.
	LastBackupTime *metav1.Time `json:"lastBackupTime,omitempty"`

	// BackupOperations contains information about recent backup operations.
	BackupOperations []BackupOperation `json:"backupOperations,omitempty"`
}

// BackupOperation defines a backup operation.
type BackupOperation struct {
	// Type specifies the type of backup operation.
	// +kubebuilder:validation:Enum=full;incremental
	Type string `json:"type"`

	// Timestamp indicates when the backup operation was performed.
	Timestamp metav1.Time `json:"timestamp"`

	// Status indicates the status of this backup operation.
	Status string `json:"status"`

	// Size indicates the size of the backup.
	Size string `json:"size,omitempty"`

	// Location indicates the location of the backup.
	Location string `json:"location,omitempty"`

	// Message provides additional information about the backup operation.
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=clusterd;cld
// +kubebuilder:printcolumn:name="Ready",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].status`,description="Shows readiness of the ClusterDeployment",priority=0
// +kubebuilder:printcolumn:name="Services",type="string",JSONPath=`.status.conditions[?(@.type=="ServicesInReadyState")].message`,description="Number of ready out of total services",priority=0
// +kubebuilder:printcolumn:name="Template",type="string",JSONPath=`.spec.template`,description="ClusterTemplate used for the ClusterDeployment",priority=0
// +kubebuilder:printcolumn:name="Messages",type="string",JSONPath=`.status.conditions[?(@.type=="Ready")].message`,description="Shows either readiness or error messages from child objects",priority=0
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`,description="Time elapsed since object creation",priority=0
// +kubebuilder:printcolumn:name="DryRun",type="string",JSONPath=`.spec.dryRun`,description="Dry Run",priority=1

// ClusterDeployment is the Schema for the ClusterDeployments API
type ClusterDeployment struct { //nolint:govet // false-positive
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterDeploymentSpec   `json:"spec,omitempty"`
	Status ClusterDeploymentStatus `json:"status,omitempty"`
}

func (in *ClusterDeployment) HelmValues() (map[string]any, error) {
	var values map[string]any

	if in.Spec.Config != nil {
		if err := yaml.Unmarshal(in.Spec.Config.Raw, &values); err != nil {
			return nil, fmt.Errorf("error unmarshalling helm values for clusterTemplate %s: %w", in.Spec.Template, err)
		}
	}

	return values, nil
}

// TODO: Refactor. Generalize default values passing. Same as in management_controller.

func (in *ClusterDeployment) SetHelmValues(values map[string]any) error {
	b, err := json.Marshal(values)
	if err != nil {
		return fmt.Errorf("error marshalling helm values for clusterTemplate %s: %w", in.Spec.Template, err)
	}

	in.Spec.Config = &apiextv1.JSON{Raw: b}
	return nil
}

func (in *ClusterDeployment) AddHelmValues(fn func(map[string]any) error) error {
	values, err := in.HelmValues()
	if err != nil {
		return err
	}

	if err := fn(values); err != nil {
		return err
	}

	return in.SetHelmValues(values)
}

func (in *ClusterDeployment) GetConditions() *[]metav1.Condition {
	return &in.Status.Conditions
}

// +kubebuilder:object:root=true

// ClusterDeploymentList contains a list of ClusterDeployment
type ClusterDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterDeployment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterDeployment{}, &ClusterDeploymentList{})
}
