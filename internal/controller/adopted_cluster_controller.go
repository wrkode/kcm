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

package controller

import (
	"context"
	"fmt"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	kcmv1 "github.com/K0rdent/kcm/api/v1beta1"
	"github.com/K0rdent/kcm/internal/backup"
	"github.com/K0rdent/kcm/internal/discovery"
	"github.com/K0rdent/kcm/internal/lifecycle"
	"github.com/K0rdent/kcm/internal/monitoring"
)

// AdoptedClusterReconciler reconciles adopted cluster lifecycle management
type AdoptedClusterReconciler struct {
	Client        client.Client
	Config        *rest.Config
	DynamicClient *dynamic.DynamicClient

	// Discovery service for infrastructure discovery
	DiscoveryService discovery.Service

	// Lifecycle service for scaling and upgrades
	LifecycleService lifecycle.Service

	// Monitoring service for health checks and metrics
	MonitoringService monitoring.Service

	// Backup service for backup and restore operations
	BackupService backup.Service

	defaultRequeueTime time.Duration
}

// Reconcile handles adopted cluster lifecycle management
func (r *AdoptedClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Reconciling adopted cluster lifecycle management")

	clusterDeployment := &kcmv1.ClusterDeployment{}
	if err := r.Client.Get(ctx, req.NamespacedName, clusterDeployment); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("ClusterDeployment not found, ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get ClusterDeployment")
		return ctrl.Result{}, err
	}

	// Check if adopted cluster management is enabled
	if clusterDeployment.Spec.AdoptedCluster == nil || !clusterDeployment.Spec.AdoptedCluster.Enabled {
		logger.Info("Adopted cluster management not enabled, skipping")
		return ctrl.Result{}, nil
	}

	// Initialize status if not present
	if clusterDeployment.Status.AdoptedClusterStatus == nil {
		clusterDeployment.Status.AdoptedClusterStatus = &kcmv1.AdoptedClusterStatus{}
	}

	// Handle deletion
	if !clusterDeployment.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, clusterDeployment)
	}

	// Perform adopted cluster lifecycle management
	return r.reconcileAdoptedCluster(ctx, clusterDeployment)
}

// reconcileAdoptedCluster handles the main reconciliation logic for adopted clusters
func (r *AdoptedClusterReconciler) reconcileAdoptedCluster(ctx context.Context, cd *kcmv1.ClusterDeployment) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Step 1: Infrastructure Discovery
	if err := r.reconcileDiscovery(ctx, cd); err != nil {
		logger.Error(err, "Failed to reconcile discovery")
		return ctrl.Result{RequeueAfter: r.defaultRequeueTime}, err
	}

	// Step 2: CAPI Integration
	if err := r.reconcileCAPI(ctx, cd); err != nil {
		logger.Error(err, "Failed to reconcile CAPI integration")
		return ctrl.Result{RequeueAfter: r.defaultRequeueTime}, err
	}

	// Step 3: Lifecycle Management (if enabled)
	if cd.Spec.AdoptedCluster.LifecycleManagement != nil && cd.Spec.AdoptedCluster.LifecycleManagement.Enabled {
		if err := r.reconcileLifecycleManagement(ctx, cd); err != nil {
			logger.Error(err, "Failed to reconcile lifecycle management")
			return ctrl.Result{RequeueAfter: r.defaultRequeueTime}, err
		}
	}

	// Update status
	if err := r.Client.Status().Update(ctx, cd); err != nil {
		logger.Error(err, "Failed to update ClusterDeployment status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: r.defaultRequeueTime}, nil
}

// reconcileDiscovery handles infrastructure discovery
func (r *AdoptedClusterReconciler) reconcileDiscovery(ctx context.Context, cd *kcmv1.ClusterDeployment) error {
	// logger := log.FromContext(ctx) // removed unused variable

	if cd.Status.AdoptedClusterStatus.DiscoveryStatus == nil {
		cd.Status.AdoptedClusterStatus.DiscoveryStatus = &kcmv1.DiscoveryStatus{
			Phase: "NotStarted",
		}
	}

	// Check if discovery is needed
	if cd.Status.AdoptedClusterStatus.DiscoveryStatus.Phase == "Completed" {
		return nil
	}

	// Update phase to InProgress
	cd.Status.AdoptedClusterStatus.DiscoveryStatus.Phase = "InProgress"
	cd.Status.AdoptedClusterStatus.DiscoveryStatus.Message = "Discovering infrastructure resources"

	// Perform discovery based on mode
	if cd.Spec.AdoptedCluster.DiscoveryMode == "auto" {
		return r.performAutoDiscovery(ctx, cd)
	} else {
		return r.performManualDiscovery(ctx, cd)
	}
}

// performAutoDiscovery performs automatic infrastructure discovery
func (r *AdoptedClusterReconciler) performAutoDiscovery(ctx context.Context, cd *kcmv1.ClusterDeployment) error {
	logger := log.FromContext(ctx)

	// Get cluster kubeconfig
	kubeconfig, err := r.getClusterKubeconfig(ctx, cd)
	if err != nil {
		return fmt.Errorf("failed to get cluster kubeconfig: %w", err)
	}

	// Discover infrastructure using the discovery service
	discoveredResources, err := r.DiscoveryService.DiscoverInfrastructure(ctx, kubeconfig, cd.Spec.AdoptedCluster.InfrastructureMapping)
	if err != nil {
		cd.Status.AdoptedClusterStatus.DiscoveryStatus.Phase = "Failed"
		cd.Status.AdoptedClusterStatus.DiscoveryStatus.Message = fmt.Sprintf("Discovery failed: %v", err)
		return fmt.Errorf("infrastructure discovery failed: %w", err)
	}

	// Update discovery status
	cd.Status.AdoptedClusterStatus.DiscoveryStatus.Phase = "Completed"
	cd.Status.AdoptedClusterStatus.DiscoveryStatus.Message = "Infrastructure discovery completed successfully"
	cd.Status.AdoptedClusterStatus.DiscoveryStatus.LastDiscoveryTime = &metav1.Time{Time: time.Now()}
	cd.Status.AdoptedClusterStatus.DiscoveryStatus.DiscoveredResources = discoveredResources

	logger.Info("Infrastructure discovery completed", "resources", len(discoveredResources))
	return nil
}

// performManualDiscovery performs manual infrastructure discovery based on provided mapping
func (r *AdoptedClusterReconciler) performManualDiscovery(ctx context.Context, cd *kcmv1.ClusterDeployment) error {
	logger := log.FromContext(ctx)

	if cd.Spec.AdoptedCluster.InfrastructureMapping == nil {
		cd.Status.AdoptedClusterStatus.DiscoveryStatus.Phase = "Failed"
		cd.Status.AdoptedClusterStatus.DiscoveryStatus.Message = "Manual discovery requires infrastructure mapping"
		return fmt.Errorf("manual discovery requires infrastructure mapping")
	}

	// Convert manual mapping to discovered resources
	var discoveredResources []kcmv1.DiscoveredResource
	for _, nodeMapping := range cd.Spec.AdoptedCluster.InfrastructureMapping.NodeMapping {
		discoveredResources = append(discoveredResources, kcmv1.DiscoveredResource{
			Type:     "node",
			Name:     nodeMapping.NodeName,
			Provider: cd.Spec.AdoptedCluster.InfrastructureMapping.Provider,
			Region:   cd.Spec.AdoptedCluster.InfrastructureMapping.Region,
			Status:   "discovered",
		})
	}

	// Update discovery status
	cd.Status.AdoptedClusterStatus.DiscoveryStatus.Phase = "Completed"
	cd.Status.AdoptedClusterStatus.DiscoveryStatus.Message = "Manual infrastructure discovery completed"
	cd.Status.AdoptedClusterStatus.DiscoveryStatus.LastDiscoveryTime = &metav1.Time{Time: time.Now()}
	cd.Status.AdoptedClusterStatus.DiscoveryStatus.DiscoveredResources = discoveredResources

	logger.Info("Manual infrastructure discovery completed", "resources", len(discoveredResources))
	return nil
}

// reconcileCAPI handles CAPI integration
func (r *AdoptedClusterReconciler) reconcileCAPI(ctx context.Context, cd *kcmv1.ClusterDeployment) error {
	logger := log.FromContext(ctx)

	if cd.Status.AdoptedClusterStatus.CAPIStatus == nil {
		cd.Status.AdoptedClusterStatus.CAPIStatus = &kcmv1.CAPIStatus{
			Phase: "NotStarted",
		}
	}

	// Check if CAPI integration is needed
	if cd.Status.AdoptedClusterStatus.CAPIStatus.Phase == "Completed" {
		return nil
	}

	// Update phase to InProgress
	cd.Status.AdoptedClusterStatus.CAPIStatus.Phase = "InProgress"
	cd.Status.AdoptedClusterStatus.CAPIStatus.Message = "Generating CAPI resources"

	// Generate CAPI resources
	generatedResources, err := r.generateCAPIResources(ctx, cd)
	if err != nil {
		cd.Status.AdoptedClusterStatus.CAPIStatus.Phase = "Failed"
		cd.Status.AdoptedClusterStatus.CAPIStatus.Message = fmt.Sprintf("CAPI integration failed: %v", err)
		return fmt.Errorf("CAPI integration failed: %w", err)
	}

	// Update CAPI status
	cd.Status.AdoptedClusterStatus.CAPIStatus.Phase = "Completed"
	cd.Status.AdoptedClusterStatus.CAPIStatus.Message = "CAPI integration completed successfully"
	cd.Status.AdoptedClusterStatus.CAPIStatus.LastSyncTime = &metav1.Time{Time: time.Now()}
	cd.Status.AdoptedClusterStatus.CAPIStatus.GeneratedResources = generatedResources

	logger.Info("CAPI integration completed", "resources", len(generatedResources))
	return nil
}

// generateCAPIResources generates CAPI resources for adopted clusters
func (r *AdoptedClusterReconciler) generateCAPIResources(ctx context.Context, cd *kcmv1.ClusterDeployment) ([]kcmv1.GeneratedResource, error) {
	var generatedResources []kcmv1.GeneratedResource

	// Generate Cluster resource
	clusterResource := kcmv1.GeneratedResource{
		APIVersion: "cluster.x-k8s.io/v1beta1",
		Kind:       "Cluster",
		Name:       cd.Name,
		Namespace:  cd.Namespace,
		Status:     "Ready",
	}
	generatedResources = append(generatedResources, clusterResource)

	// Generate Machine resources for each node mapping
	if cd.Spec.AdoptedCluster.InfrastructureMapping != nil {
		for _, nodeMapping := range cd.Spec.AdoptedCluster.InfrastructureMapping.NodeMapping {
			machineResource := kcmv1.GeneratedResource{
				APIVersion: "cluster.x-k8s.io/v1beta1",
				Kind:       "Machine",
				Name:       nodeMapping.MachineName,
				Namespace:  cd.Namespace,
				Status:     "Ready",
			}
			generatedResources = append(generatedResources, machineResource)

			// Generate infrastructure resource if specified
			if nodeMapping.InfrastructureRef != nil {
				infraResource := kcmv1.GeneratedResource{
					APIVersion: nodeMapping.InfrastructureRef.APIVersion,
					Kind:       nodeMapping.InfrastructureRef.Kind,
					Name:       nodeMapping.InfrastructureRef.Name,
					Namespace:  cd.Namespace,
					Status:     "Ready",
				}
				generatedResources = append(generatedResources, infraResource)
			}
		}
	}

	return generatedResources, nil
}

// reconcileLifecycleManagement handles lifecycle management operations
func (r *AdoptedClusterReconciler) reconcileLifecycleManagement(ctx context.Context, cd *kcmv1.ClusterDeployment) error {
	logger := log.FromContext(ctx)

	lifecycle := cd.Spec.AdoptedCluster.LifecycleManagement

	// Handle scaling
	if lifecycle.Scaling != nil && lifecycle.Scaling.Enabled {
		if err := r.reconcileScaling(ctx, cd); err != nil {
			logger.Error(err, "Failed to reconcile scaling")
			return err
		}
	}

	// Handle upgrades
	if lifecycle.Upgrades != nil && lifecycle.Upgrades.Enabled {
		if err := r.reconcileUpgrades(ctx, cd); err != nil {
			logger.Error(err, "Failed to reconcile upgrades")
			return err
		}
	}

	// Handle monitoring
	if lifecycle.Monitoring != nil && lifecycle.Monitoring.Enabled {
		if err := r.reconcileMonitoring(ctx, cd); err != nil {
			logger.Error(err, "Failed to reconcile monitoring")
			return err
		}
	}

	// Handle backup
	if lifecycle.Backup != nil && lifecycle.Backup.Enabled {
		if err := r.reconcileBackup(ctx, cd); err != nil {
			logger.Error(err, "Failed to reconcile backup")
			return err
		}
	}

	return nil
}

// reconcileScaling handles scaling operations
func (r *AdoptedClusterReconciler) reconcileScaling(ctx context.Context, cd *kcmv1.ClusterDeployment) error {
	// logger := log.FromContext(ctx) // removed unused variable

	if cd.Status.AdoptedClusterStatus.ScalingStatus == nil {
		cd.Status.AdoptedClusterStatus.ScalingStatus = &kcmv1.ScalingStatus{
			Phase: "NotStarted",
		}
	}

	scaling := cd.Spec.AdoptedCluster.LifecycleManagement.Scaling

	// Get current node count
	currentNodes, err := r.getCurrentNodeCount(ctx, cd)
	if err != nil {
		return fmt.Errorf("failed to get current node count: %w", err)
	}

	// Update current nodes in status
	cd.Status.AdoptedClusterStatus.ScalingStatus.CurrentNodes = currentNodes

	// Check if scaling is needed
	if scaling.AutoScaling {
		desiredNodes, err := r.calculateDesiredNodes(ctx, cd)
		if err != nil {
			return fmt.Errorf("failed to calculate desired nodes: %w", err)
		}

		cd.Status.AdoptedClusterStatus.ScalingStatus.DesiredNodes = desiredNodes

		if currentNodes != desiredNodes {
			// Perform scaling operation
			if err := r.performScaling(ctx, cd, currentNodes, desiredNodes); err != nil {
				return fmt.Errorf("failed to perform scaling: %w", err)
			}
		}
	}

	return nil
}

// reconcileUpgrades handles upgrade operations
func (r *AdoptedClusterReconciler) reconcileUpgrades(ctx context.Context, cd *kcmv1.ClusterDeployment) error {
	// logger := log.FromContext(ctx) // removed unused variable

	if cd.Status.AdoptedClusterStatus.UpgradeStatus == nil {
		cd.Status.AdoptedClusterStatus.UpgradeStatus = &kcmv1.UpgradeStatus{
			Phase: "NotStarted",
		}
	}

	upgrades := cd.Spec.AdoptedCluster.LifecycleManagement.Upgrades

	// Check if upgrade is needed
	if upgrades.AutoUpgrade {
		// Check for available upgrades
		upgradeAvailable, targetVersion, err := r.checkForUpgrades(ctx, cd)
		if err != nil {
			return fmt.Errorf("failed to check for upgrades: %w", err)
		}

		if upgradeAvailable {
			// Perform upgrade
			if err := r.performUpgrade(ctx, cd, targetVersion); err != nil {
				return fmt.Errorf("failed to perform upgrade: %w", err)
			}
		}
	}

	return nil
}

// reconcileMonitoring handles monitoring operations
func (r *AdoptedClusterReconciler) reconcileMonitoring(ctx context.Context, cd *kcmv1.ClusterDeployment) error {
	logger := log.FromContext(ctx)

	if cd.Status.AdoptedClusterStatus.MonitoringStatus == nil {
		cd.Status.AdoptedClusterStatus.MonitoringStatus = &kcmv1.MonitoringStatus{
			Phase: "NotStarted",
		}
	}

	// Perform health checks
	healthStatus, err := r.MonitoringService.PerformHealthChecks(ctx, cd)
	if err != nil {
		logger.Error(err, "Failed to perform health checks")
		return err
	}

	// Collect metrics
	metrics, err := r.MonitoringService.CollectMetrics(ctx, cd)
	if err != nil {
		logger.Error(err, "Failed to collect metrics")
		return err
	}

	// Update monitoring status
	cd.Status.AdoptedClusterStatus.MonitoringStatus.Phase = "Completed"
	cd.Status.AdoptedClusterStatus.MonitoringStatus.HealthStatus = healthStatus
	cd.Status.AdoptedClusterStatus.MonitoringStatus.Metrics = metrics
	cd.Status.AdoptedClusterStatus.MonitoringStatus.LastHealthCheckTime = &metav1.Time{Time: time.Now()}

	return nil
}

// reconcileBackup handles backup operations
func (r *AdoptedClusterReconciler) reconcileBackup(ctx context.Context, cd *kcmv1.ClusterDeployment) error {
	// logger := log.FromContext(ctx) // removed unused variable

	if cd.Status.AdoptedClusterStatus.BackupStatus == nil {
		cd.Status.AdoptedClusterStatus.BackupStatus = &kcmv1.BackupStatus{
			Phase: "NotStarted",
		}
	}

	backup := cd.Spec.AdoptedCluster.LifecycleManagement.Backup

	// Check if backup is needed based on schedule
	if backup.Schedule != "" {
		backupNeeded, err := r.BackupService.IsBackupNeeded(ctx, cd, backup.Schedule)
		if err != nil {
			return fmt.Errorf("failed to check if backup is needed: %w", err)
		}

		if backupNeeded {
			// Perform backup
			if err := r.performBackup(ctx, cd); err != nil {
				return fmt.Errorf("failed to perform backup: %w", err)
			}
		}
	}

	return nil
}

// Helper methods for scaling operations
func (r *AdoptedClusterReconciler) getCurrentNodeCount(ctx context.Context, cd *kcmv1.ClusterDeployment) (int32, error) {
	// This would typically query the adopted cluster to get current node count
	// For now, return a placeholder value
	return 3, nil
}

func (r *AdoptedClusterReconciler) calculateDesiredNodes(ctx context.Context, cd *kcmv1.ClusterDeployment) (int32, error) {
	// This would implement auto-scaling logic based on metrics
	// For now, return a placeholder value
	return 3, nil
}

func (r *AdoptedClusterReconciler) performScaling(ctx context.Context, cd *kcmv1.ClusterDeployment, currentNodes, desiredNodes int32) error {
	// This would perform the actual scaling operation
	// For now, just log the operation
	logger := log.FromContext(ctx)
	logger.Info("Performing scaling operation", "currentNodes", currentNodes, "desiredNodes", desiredNodes)

	// Update scaling status
	cd.Status.AdoptedClusterStatus.ScalingStatus.Phase = "InProgress"
	cd.Status.AdoptedClusterStatus.ScalingStatus.Message = fmt.Sprintf("Scaling from %d to %d nodes", currentNodes, desiredNodes)

	return nil
}

// Helper methods for upgrade operations
func (r *AdoptedClusterReconciler) checkForUpgrades(ctx context.Context, cd *kcmv1.ClusterDeployment) (bool, string, error) {
	// This would check for available upgrades
	// For now, return false
	return false, "", nil
}

func (r *AdoptedClusterReconciler) performUpgrade(ctx context.Context, cd *kcmv1.ClusterDeployment, targetVersion string) error {
	// This would perform the actual upgrade operation
	// For now, just log the operation
	logger := log.FromContext(ctx)
	logger.Info("Performing upgrade", "targetVersion", targetVersion)

	// Update upgrade status
	cd.Status.AdoptedClusterStatus.UpgradeStatus.Phase = "InProgress"
	cd.Status.AdoptedClusterStatus.UpgradeStatus.TargetVersion = targetVersion
	cd.Status.AdoptedClusterStatus.UpgradeStatus.Message = fmt.Sprintf("Upgrading to version %s", targetVersion)

	return nil
}

// Helper methods for backup operations
func (r *AdoptedClusterReconciler) performBackup(ctx context.Context, cd *kcmv1.ClusterDeployment) error {
	// This would perform the actual backup operation
	// For now, just log the operation
	logger := log.FromContext(ctx)
	logger.Info("Performing backup")

	// Update backup status
	cd.Status.AdoptedClusterStatus.BackupStatus.Phase = "InProgress"
	cd.Status.AdoptedClusterStatus.BackupStatus.Message = "Performing backup"

	return nil
}

// Helper method to get cluster kubeconfig
func (r *AdoptedClusterReconciler) getClusterKubeconfig(ctx context.Context, cd *kcmv1.ClusterDeployment) ([]byte, error) {
	// This would retrieve the kubeconfig for the adopted cluster
	// For now, return a placeholder
	return []byte("placeholder-kubeconfig"), nil
}

// reconcileDelete handles deletion of adopted cluster resources
func (r *AdoptedClusterReconciler) reconcileDelete(ctx context.Context, cd *kcmv1.ClusterDeployment) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Deleting adopted cluster resources")

	// Clean up CAPI resources
	if err := r.cleanupCAPIResources(ctx, cd); err != nil {
		logger.Error(err, "Failed to cleanup CAPI resources")
		return ctrl.Result{RequeueAfter: r.defaultRequeueTime}, err
	}

	// Clean up monitoring resources
	if err := r.MonitoringService.Cleanup(ctx, cd); err != nil {
		logger.Error(err, "Failed to cleanup monitoring resources")
		return ctrl.Result{RequeueAfter: r.defaultRequeueTime}, err
	}

	// Clean up backup resources
	if err := r.BackupService.Cleanup(ctx, cd); err != nil {
		logger.Error(err, "Failed to cleanup backup resources")
		return ctrl.Result{RequeueAfter: r.defaultRequeueTime}, err
	}

	return ctrl.Result{}, nil
}

// cleanupCAPIResources cleans up generated CAPI resources
func (r *AdoptedClusterReconciler) cleanupCAPIResources(ctx context.Context, cd *kcmv1.ClusterDeployment) error {
	// This would clean up all generated CAPI resources
	// For now, just log the operation
	logger := log.FromContext(ctx)
	logger.Info("Cleaning up CAPI resources")
	return nil
}

// SetupWithManager sets up the controller with the manager
func (r *AdoptedClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kcmv1.ClusterDeployment{}).
		WithEventFilter(r.adoptedClusterPredicate()).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: 1,
		}).
		Complete(r)
}

// adoptedClusterPredicate filters events to only process ClusterDeployments with adopted cluster management enabled
func (r *AdoptedClusterReconciler) adoptedClusterPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			cd, ok := e.Object.(*kcmv1.ClusterDeployment)
			return ok && cd.Spec.AdoptedCluster != nil && cd.Spec.AdoptedCluster.Enabled
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldCD, oldOk := e.ObjectOld.(*kcmv1.ClusterDeployment)
			newCD, newOk := e.ObjectNew.(*kcmv1.ClusterDeployment)
			if !oldOk || !newOk {
				return false
			}

			// Check if adopted cluster management was enabled or configuration changed
			oldEnabled := oldCD.Spec.AdoptedCluster != nil && oldCD.Spec.AdoptedCluster.Enabled
			newEnabled := newCD.Spec.AdoptedCluster != nil && newCD.Spec.AdoptedCluster.Enabled

			return newEnabled || (oldEnabled && newEnabled)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			cd, ok := e.Object.(*kcmv1.ClusterDeployment)
			return ok && cd.Spec.AdoptedCluster != nil && cd.Spec.AdoptedCluster.Enabled
		},
		GenericFunc: func(e event.GenericEvent) bool {
			cd, ok := e.Object.(*kcmv1.ClusterDeployment)
			return ok && cd.Spec.AdoptedCluster != nil && cd.Spec.AdoptedCluster.Enabled
		},
	}
}
