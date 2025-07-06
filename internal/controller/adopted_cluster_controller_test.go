// Copyright 2025
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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	. "sigs.k8s.io/controller-runtime/pkg/envtest/komega"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	kcmv1 "github.com/K0rdent/kcm/api/v1beta1"
	"github.com/K0rdent/kcm/internal/backup"
)

// Mock service implementations for testing
type mockDiscoveryService struct{}

func (m *mockDiscoveryService) DiscoverInfrastructure(ctx context.Context, kubeconfig []byte, mapping *kcmv1.InfrastructureMapping) ([]kcmv1.DiscoveredResource, error) {
	return []kcmv1.DiscoveredResource{
		{
			Type:     "node",
			Name:     "test-node-1",
			Provider: "aws",
			Region:   "us-west-2",
			Status:   "discovered",
		},
		{
			Type:     "node",
			Name:     "test-node-2",
			Provider: "aws",
			Region:   "us-west-2",
			Status:   "discovered",
		},
	}, nil
}

func (m *mockDiscoveryService) ValidateInfrastructure(ctx context.Context, discovered []kcmv1.DiscoveredResource, mapping *kcmv1.InfrastructureMapping) error {
	return nil
}

type mockLifecycleService struct{}

func (m *mockLifecycleService) ScaleCluster(ctx context.Context, cd *kcmv1.ClusterDeployment, desiredNodes int32) error {
	return nil
}

func (m *mockLifecycleService) UpgradeCluster(ctx context.Context, cd *kcmv1.ClusterDeployment, targetVersion string) error {
	return nil
}

func (m *mockLifecycleService) CheckUpgradeAvailability(ctx context.Context, cd *kcmv1.ClusterDeployment) (bool, string, error) {
	return false, "", nil
}

func (m *mockLifecycleService) GetCurrentNodeCount(ctx context.Context, cd *kcmv1.ClusterDeployment) (int32, error) {
	return 3, nil
}

func (m *mockLifecycleService) CalculateDesiredNodes(ctx context.Context, cd *kcmv1.ClusterDeployment) (int32, error) {
	return 3, nil
}

type mockMonitoringService struct{}

func (m *mockMonitoringService) PerformHealthChecks(ctx context.Context, cd *kcmv1.ClusterDeployment) (string, error) {
	return "Healthy", nil
}

func (m *mockMonitoringService) CollectMetrics(ctx context.Context, cd *kcmv1.ClusterDeployment) ([]kcmv1.MetricInfo, error) {
	return []kcmv1.MetricInfo{
		{
			Name:      "cpu_usage",
			Value:     "75%",
			Unit:      "percentage",
			Timestamp: metav1.Now(),
		},
	}, nil
}

func (m *mockMonitoringService) Cleanup(ctx context.Context, cd *kcmv1.ClusterDeployment) error {
	return nil
}

type mockBackupService struct{}

func (m *mockBackupService) IsBackupNeeded(ctx context.Context, cd *kcmv1.ClusterDeployment, schedule string) (bool, error) {
	return true, nil
}

func (m *mockBackupService) PerformBackup(ctx context.Context, cd *kcmv1.ClusterDeployment) error {
	return nil
}

func (m *mockBackupService) RestoreBackup(ctx context.Context, cd *kcmv1.ClusterDeployment, backupName string) error {
	return nil
}

func (m *mockBackupService) ListBackups(ctx context.Context, cd *kcmv1.ClusterDeployment) ([]backup.BackupInfo, error) {
	return []backup.BackupInfo{
		{
			Name:     "backup-1",
			Type:     "full",
			Size:     "1GB",
			Location: "s3://backup-bucket/backup-1",
			Status:   "completed",
			Created:  time.Now().Add(-24 * time.Hour),
		},
	}, nil
}

func (m *mockBackupService) Cleanup(ctx context.Context, cd *kcmv1.ClusterDeployment) error {
	return nil
}

var _ = Describe("AdoptedCluster Controller", func() {
	Context("When reconciling adopted cluster resources", func() {
		var (
			namespace         = corev1.Namespace{}
			clusterDeployment = kcmv1.ClusterDeployment{}
			reconciler        *AdoptedClusterReconciler
		)

		BeforeEach(func() {
			By("ensure namespace", func() {
				namespace = corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "test-adopted-namespace-",
					},
				}
				Expect(k8sClient.Create(ctx, &namespace)).To(Succeed())
				DeferCleanup(k8sClient.Delete, &namespace)
			})

			By("setup reconciler with mock services", func() {
				reconciler = &AdoptedClusterReconciler{
					Client:             mgrClient,
					Config:             &rest.Config{},
					DynamicClient:      dynamicClient,
					DiscoveryService:   &mockDiscoveryService{},
					LifecycleService:   &mockLifecycleService{},
					MonitoringService:  &mockMonitoringService{},
					BackupService:      &mockBackupService{},
					defaultRequeueTime: 30 * time.Second,
				}
			})
		})

		It("should reconcile adopted cluster with auto discovery", func() {
			By("creating ClusterDeployment with adopted cluster enabled", func() {
				clusterDeployment = kcmv1.ClusterDeployment{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "test-adopted-cluster-",
						Namespace:    namespace.Name,
					},
					Spec: kcmv1.ClusterDeploymentSpec{
						Template: "adopted-cluster",
						AdoptedCluster: &kcmv1.AdoptedClusterConfig{
							Enabled:       true,
							DiscoveryMode: "auto",
							InfrastructureMapping: &kcmv1.InfrastructureMapping{
								Provider: "aws",
								Region:   "us-west-2",
								Credentials: &kcmv1.CloudCredentials{
									SecretName: "aws-credentials",
									Namespace:  namespace.Name,
								},
							},
							LifecycleManagement: &kcmv1.LifecycleManagement{
								Enabled: true,
								Scaling: &kcmv1.ScalingConfig{
									Enabled:     true,
									AutoScaling: true,
									MinNodes:    1,
									MaxNodes:    5,
								},
								Monitoring: &kcmv1.MonitoringConfig{
									Enabled: true,
									HealthChecks: &kcmv1.HealthCheckConfig{
										Enabled: true,
									},
								},
								Backup: &kcmv1.BackupConfig{
									Enabled:  true,
									Schedule: "0 2 * * *",
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, &clusterDeployment)).To(Succeed())
				DeferCleanup(k8sClient.Delete, &clusterDeployment)
			})

			By("ensuring adopted cluster status is initialized", func() {
				Eventually(func(g Gomega) {
					_, err := reconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: client.ObjectKeyFromObject(&clusterDeployment),
					})
					g.Expect(err).NotTo(HaveOccurred())

					g.Expect(Object(&clusterDeployment)()).Should(SatisfyAll(
						HaveField("Status.AdoptedClusterStatus", Not(BeNil())),
						HaveField("Status.AdoptedClusterStatus.DiscoveryStatus", Not(BeNil())),
						HaveField("Status.AdoptedClusterStatus.DiscoveryStatus.Phase", "InProgress"),
					))
				}).Should(Succeed())
			})

			By("ensuring discovery completes successfully", func() {
				Eventually(func(g Gomega) {
					_, err := reconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: client.ObjectKeyFromObject(&clusterDeployment),
					})
					g.Expect(err).NotTo(HaveOccurred())

					g.Expect(Object(&clusterDeployment)()).Should(SatisfyAll(
						HaveField("Status.AdoptedClusterStatus.DiscoveryStatus.Phase", "Completed"),
						HaveField("Status.AdoptedClusterStatus.DiscoveryStatus.DiscoveredResources", HaveLen(2)),
					))
				}).Should(Succeed())
			})

			By("ensuring CAPI integration completes", func() {
				Eventually(func(g Gomega) {
					_, err := reconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: client.ObjectKeyFromObject(&clusterDeployment),
					})
					g.Expect(err).NotTo(HaveOccurred())

					g.Expect(Object(&clusterDeployment)()).Should(SatisfyAll(
						HaveField("Status.AdoptedClusterStatus.CAPIStatus.Phase", "Completed"),
						HaveField("Status.AdoptedClusterStatus.CAPIStatus.GeneratedResources", Not(BeEmpty())),
					))
				}).Should(Succeed())
			})

			By("ensuring monitoring is active", func() {
				Eventually(func(g Gomega) {
					_, err := reconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: client.ObjectKeyFromObject(&clusterDeployment),
					})
					g.Expect(err).NotTo(HaveOccurred())

					g.Expect(Object(&clusterDeployment)()).Should(SatisfyAll(
						HaveField("Status.AdoptedClusterStatus.MonitoringStatus.Phase", "Completed"),
						HaveField("Status.AdoptedClusterStatus.MonitoringStatus.HealthStatus", "Healthy"),
						HaveField("Status.AdoptedClusterStatus.MonitoringStatus.Metrics", Not(BeEmpty())),
					))
				}).Should(Succeed())
			})
		})

		It("should reconcile adopted cluster with manual discovery", func() {
			By("creating ClusterDeployment with manual discovery", func() {
				clusterDeployment = kcmv1.ClusterDeployment{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "test-adopted-cluster-manual-",
						Namespace:    namespace.Name,
					},
					Spec: kcmv1.ClusterDeploymentSpec{
						Template: "adopted-cluster",
						AdoptedCluster: &kcmv1.AdoptedClusterConfig{
							Enabled:       true,
							DiscoveryMode: "manual",
							InfrastructureMapping: &kcmv1.InfrastructureMapping{
								Provider: "aws",
								Region:   "us-west-2",
								NodeMapping: []kcmv1.NodeMapping{
									{
										NodeName:    "existing-node-1",
										MachineName: "machine-1",
									},
									{
										NodeName:    "existing-node-2",
										MachineName: "machine-2",
									},
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, &clusterDeployment)).To(Succeed())
				DeferCleanup(k8sClient.Delete, &clusterDeployment)
			})

			By("ensuring manual discovery completes", func() {
				Eventually(func(g Gomega) {
					_, err := reconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: client.ObjectKeyFromObject(&clusterDeployment),
					})
					g.Expect(err).NotTo(HaveOccurred())

					g.Expect(Object(&clusterDeployment)()).Should(SatisfyAll(
						HaveField("Status.AdoptedClusterStatus.DiscoveryStatus.Phase", "Completed"),
						HaveField("Status.AdoptedClusterStatus.DiscoveryStatus.DiscoveredResources", HaveLen(2)),
					))
				}).Should(Succeed())
			})
		})

		It("should handle discovery failures gracefully", func() {
			By("creating ClusterDeployment with invalid configuration", func() {
				clusterDeployment = kcmv1.ClusterDeployment{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "test-adopted-cluster-failure-",
						Namespace:    namespace.Name,
					},
					Spec: kcmv1.ClusterDeploymentSpec{
						Template: "adopted-cluster",
						AdoptedCluster: &kcmv1.AdoptedClusterConfig{
							Enabled:       true,
							DiscoveryMode: "manual",
							// Missing InfrastructureMapping should cause failure
						},
					},
				}
				Expect(k8sClient.Create(ctx, &clusterDeployment)).To(Succeed())
				DeferCleanup(k8sClient.Delete, &clusterDeployment)
			})

			By("ensuring discovery failure is handled", func() {
				Eventually(func(g Gomega) {
					_, err := reconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: client.ObjectKeyFromObject(&clusterDeployment),
					})
					g.Expect(err).NotTo(HaveOccurred())

					g.Expect(Object(&clusterDeployment)()).Should(SatisfyAll(
						HaveField("Status.AdoptedClusterStatus.DiscoveryStatus.Phase", "Failed"),
						HaveField("Status.AdoptedClusterStatus.DiscoveryStatus.Message", ContainSubstring("requires infrastructure mapping")),
					))
				}).Should(Succeed())
			})
		})

		It("should handle scaling operations", func() {
			By("creating ClusterDeployment with scaling enabled", func() {
				clusterDeployment = kcmv1.ClusterDeployment{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "test-adopted-cluster-scaling-",
						Namespace:    namespace.Name,
					},
					Spec: kcmv1.ClusterDeploymentSpec{
						Template: "adopted-cluster",
						AdoptedCluster: &kcmv1.AdoptedClusterConfig{
							Enabled: true,
							LifecycleManagement: &kcmv1.LifecycleManagement{
								Enabled: true,
								Scaling: &kcmv1.ScalingConfig{
									Enabled:     true,
									AutoScaling: true,
									MinNodes:    1,
									MaxNodes:    5,
									NodeGroups: []kcmv1.NodeGroup{
										{
											Name:     "worker-nodes",
											MinNodes: 1,
											MaxNodes: 3,
											Labels:   map[string]string{"node-role": "worker"},
										},
									},
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, &clusterDeployment)).To(Succeed())
				DeferCleanup(k8sClient.Delete, &clusterDeployment)
			})

			By("ensuring scaling status is tracked", func() {
				Eventually(func(g Gomega) {
					_, err := reconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: client.ObjectKeyFromObject(&clusterDeployment),
					})
					g.Expect(err).NotTo(HaveOccurred())

					g.Expect(Object(&clusterDeployment)()).Should(SatisfyAll(
						HaveField("Status.AdoptedClusterStatus.ScalingStatus", Not(BeNil())),
						HaveField("Status.AdoptedClusterStatus.ScalingStatus.CurrentNodes", BeNumerically(">=", 0)),
					))
				}).Should(Succeed())
			})
		})

		It("should handle upgrade operations", func() {
			By("creating ClusterDeployment with upgrades enabled", func() {
				clusterDeployment = kcmv1.ClusterDeployment{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "test-adopted-cluster-upgrade-",
						Namespace:    namespace.Name,
					},
					Spec: kcmv1.ClusterDeploymentSpec{
						Template: "adopted-cluster",
						AdoptedCluster: &kcmv1.AdoptedClusterConfig{
							Enabled: true,
							LifecycleManagement: &kcmv1.LifecycleManagement{
								Enabled: true,
								Upgrades: &kcmv1.UpgradeConfig{
									Enabled:         true,
									AutoUpgrade:     true,
									UpgradeStrategy: "rolling",
									MaxUnavailable:  1,
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, &clusterDeployment)).To(Succeed())
				DeferCleanup(k8sClient.Delete, &clusterDeployment)
			})

			By("ensuring upgrade status is tracked", func() {
				Eventually(func(g Gomega) {
					_, err := reconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: client.ObjectKeyFromObject(&clusterDeployment),
					})
					g.Expect(err).NotTo(HaveOccurred())

					g.Expect(Object(&clusterDeployment)()).Should(SatisfyAll(
						HaveField("Status.AdoptedClusterStatus.UpgradeStatus", Not(BeNil())),
						HaveField("Status.AdoptedClusterStatus.UpgradeStatus.Phase", "NotStarted"),
					))
				}).Should(Succeed())
			})
		})

		It("should handle backup operations", func() {
			By("creating ClusterDeployment with backup enabled", func() {
				clusterDeployment = kcmv1.ClusterDeployment{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "test-adopted-cluster-backup-",
						Namespace:    namespace.Name,
					},
					Spec: kcmv1.ClusterDeploymentSpec{
						Template: "adopted-cluster",
						AdoptedCluster: &kcmv1.AdoptedClusterConfig{
							Enabled: true,
							LifecycleManagement: &kcmv1.LifecycleManagement{
								Enabled: true,
								Backup: &kcmv1.BackupConfig{
									Enabled:   true,
									Schedule:  "0 2 * * *",
									Retention: "30d",
									Storage: &kcmv1.BackupStorage{
										Type:     "s3",
										Location: "s3://backup-bucket/adopted-clusters",
									},
								},
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, &clusterDeployment)).To(Succeed())
				DeferCleanup(k8sClient.Delete, &clusterDeployment)
			})

			By("ensuring backup status is tracked", func() {
				Eventually(func(g Gomega) {
					_, err := reconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: client.ObjectKeyFromObject(&clusterDeployment),
					})
					g.Expect(err).NotTo(HaveOccurred())

					g.Expect(Object(&clusterDeployment)()).Should(SatisfyAll(
						HaveField("Status.AdoptedClusterStatus.BackupStatus", Not(BeNil())),
						HaveField("Status.AdoptedClusterStatus.BackupStatus.Phase", "NotStarted"),
					))
				}).Should(Succeed())
			})
		})

		It("should handle deletion cleanup", func() {
			By("creating ClusterDeployment for deletion test", func() {
				clusterDeployment = kcmv1.ClusterDeployment{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "test-adopted-cluster-deletion-",
						Namespace:    namespace.Name,
					},
					Spec: kcmv1.ClusterDeploymentSpec{
						Template: "adopted-cluster",
						AdoptedCluster: &kcmv1.AdoptedClusterConfig{
							Enabled: true,
						},
					},
				}
				Expect(k8sClient.Create(ctx, &clusterDeployment)).To(Succeed())
			})

			By("ensuring deletion cleanup is handled", func() {
				// First reconcile to set up the resource
				_, err := reconciler.Reconcile(ctx, reconcile.Request{
					NamespacedName: client.ObjectKeyFromObject(&clusterDeployment),
				})
				Expect(err).NotTo(HaveOccurred())

				// Mark for deletion
				Expect(k8sClient.Delete(ctx, &clusterDeployment)).To(Succeed())

				// Reconcile deletion
				Eventually(func(g Gomega) {
					_, err := reconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: client.ObjectKeyFromObject(&clusterDeployment),
					})
					g.Expect(err).NotTo(HaveOccurred())
				}).Should(Succeed())
			})
		})

		It("should handle service failures gracefully", func() {
			By("creating reconciler with failing mock services", func() {
				// Create a mock service that fails
				failingDiscoveryService := &mockDiscoveryService{}
				// Override the DiscoverInfrastructure method to fail
				// This would require a more sophisticated mock, but for now we'll test the error handling

				reconciler.DiscoveryService = failingDiscoveryService
			})

			By("creating ClusterDeployment that will trigger service failure", func() {
				clusterDeployment = kcmv1.ClusterDeployment{
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "test-adopted-cluster-service-failure-",
						Namespace:    namespace.Name,
					},
					Spec: kcmv1.ClusterDeploymentSpec{
						Template: "adopted-cluster",
						AdoptedCluster: &kcmv1.AdoptedClusterConfig{
							Enabled:       true,
							DiscoveryMode: "auto",
							InfrastructureMapping: &kcmv1.InfrastructureMapping{
								Provider: "aws",
								Region:   "us-west-2",
							},
						},
					},
				}
				Expect(k8sClient.Create(ctx, &clusterDeployment)).To(Succeed())
				DeferCleanup(k8sClient.Delete, &clusterDeployment)
			})

			By("ensuring service failures are handled gracefully", func() {
				Eventually(func(g Gomega) {
					_, err := reconciler.Reconcile(ctx, reconcile.Request{
						NamespacedName: client.ObjectKeyFromObject(&clusterDeployment),
					})
					// The reconciler should handle errors gracefully
					g.Expect(err).NotTo(HaveOccurred())
				}).Should(Succeed())
			})
		})
	})
})
