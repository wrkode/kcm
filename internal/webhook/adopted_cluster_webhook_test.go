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

package webhook

import (
	"testing"

	. "github.com/onsi/gomega"
	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	kcmv1 "github.com/K0rdent/kcm/api/v1beta1"
	"github.com/K0rdent/kcm/test/objects/clusterdeployment"
	"github.com/K0rdent/kcm/test/objects/credential"
	"github.com/K0rdent/kcm/test/objects/management"
	"github.com/K0rdent/kcm/test/objects/providerinterface"
	"github.com/K0rdent/kcm/test/objects/template"
	"github.com/K0rdent/kcm/test/scheme"
)

var (
	adoptedTestTemplateName   = "adopted-cluster-template"
	adoptedTestCredentialName = "aws-credential"
	adoptedTestNamespace      = "test"

	adoptedMgmt = management.NewManagement(
		management.WithAvailableProviders(kcmv1.Providers{
			"infrastructure-aws",
			"control-plane-k0smotron",
			"bootstrap-k0smotron",
		}),
	)

	adoptedCred = credential.NewCredential(
		credential.WithName(adoptedTestCredentialName),
		credential.WithReady(true),
		credential.WithIdentityRef(
			&corev1.ObjectReference{
				Kind: "AWSClusterStaticIdentity",
				Name: "awsclid",
			}),
	)

	adoptedProviderInterface = providerinterface.NewAWSProviderInterface()
)

func TestAdoptedClusterValidateCreate(t *testing.T) {
	ctx := admission.NewContextWithRequest(t.Context(), admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: admissionv1.Create,
		},
	})

	tests := []struct {
		name              string
		ClusterDeployment *kcmv1.ClusterDeployment
		existingObjects   []runtime.Object
		err               string
		warnings          admission.Warnings
	}{
		{
			name: "should pass with valid adopted cluster configuration",
			ClusterDeployment: clusterdeployment.NewClusterDeployment(
				clusterdeployment.WithClusterTemplate(adoptedTestTemplateName),
				clusterdeployment.WithCredential(adoptedTestCredentialName),
				clusterdeployment.WithAdoptedCluster(&kcmv1.AdoptedClusterConfig{
					Enabled:       true,
					DiscoveryMode: "auto",
					InfrastructureMapping: &kcmv1.InfrastructureMapping{
						Provider: "aws",
						Region:   "us-west-2",
						Credentials: &kcmv1.CloudCredentials{
							SecretName: "aws-credentials",
							Namespace:  adoptedTestNamespace,
						},
					},
				}),
			),
			existingObjects: []runtime.Object{
				adoptedMgmt,
				adoptedCred,
				adoptedProviderInterface,
				template.NewClusterTemplate(
					template.WithName(adoptedTestTemplateName),
					template.WithProvidersStatus(
						"infrastructure-aws",
						"control-plane-k0smotron",
						"bootstrap-k0smotron",
					),
					template.WithValidationStatus(kcmv1.TemplateValidationStatus{Valid: true}),
				),
			},
		},
		{
			name: "should fail with adopted cluster enabled but no infrastructure mapping",
			ClusterDeployment: clusterdeployment.NewClusterDeployment(
				clusterdeployment.WithClusterTemplate(testTemplateName),
				clusterdeployment.WithCredential(testCredentialName),
				clusterdeployment.WithAdoptedCluster(&kcmv1.AdoptedClusterConfig{
					Enabled:       true,
					DiscoveryMode: "auto",
					// Missing InfrastructureMapping
				}),
			),
			existingObjects: []runtime.Object{
				mgmt,
				cred,
				providerInterface,
				template.NewClusterTemplate(
					template.WithName(testTemplateName),
					template.WithProvidersStatus(
						"infrastructure-aws",
						"control-plane-k0smotron",
						"bootstrap-k0smotron",
					),
					template.WithValidationStatus(kcmv1.TemplateValidationStatus{Valid: true}),
				),
			},
			err: "adopted cluster configuration requires infrastructure mapping",
		},
		{
			name: "should fail with invalid discovery mode",
			ClusterDeployment: clusterdeployment.NewClusterDeployment(
				clusterdeployment.WithClusterTemplate(testTemplateName),
				clusterdeployment.WithCredential(testCredentialName),
				clusterdeployment.WithAdoptedCluster(&kcmv1.AdoptedClusterConfig{
					Enabled:       true,
					DiscoveryMode: "invalid-mode",
					InfrastructureMapping: &kcmv1.InfrastructureMapping{
						Provider: "aws",
						Region:   "us-west-2",
					},
				}),
			),
			existingObjects: []runtime.Object{
				mgmt,
				cred,
				providerInterface,
				template.NewClusterTemplate(
					template.WithName(testTemplateName),
					template.WithProvidersStatus(
						"infrastructure-aws",
						"control-plane-k0smotron",
						"bootstrap-k0smotron",
					),
					template.WithValidationStatus(kcmv1.TemplateValidationStatus{Valid: true}),
				),
			},
			err: "invalid discovery mode: invalid-mode. Must be 'auto' or 'manual'",
		},
		{
			name: "should fail with manual discovery but no node mapping",
			ClusterDeployment: clusterdeployment.NewClusterDeployment(
				clusterdeployment.WithClusterTemplate(testTemplateName),
				clusterdeployment.WithCredential(testCredentialName),
				clusterdeployment.WithAdoptedCluster(&kcmv1.AdoptedClusterConfig{
					Enabled:       true,
					DiscoveryMode: "manual",
					InfrastructureMapping: &kcmv1.InfrastructureMapping{
						Provider: "aws",
						Region:   "us-west-2",
						// Missing NodeMapping for manual discovery
					},
				}),
			),
			existingObjects: []runtime.Object{
				mgmt,
				cred,
				providerInterface,
				template.NewClusterTemplate(
					template.WithName(testTemplateName),
					template.WithProvidersStatus(
						"infrastructure-aws",
						"control-plane-k0smotron",
						"bootstrap-k0smotron",
					),
					template.WithValidationStatus(kcmv1.TemplateValidationStatus{Valid: true}),
				),
			},
			err: "manual discovery mode requires node mapping configuration",
		},
		{
			name: "should fail with invalid provider in infrastructure mapping",
			ClusterDeployment: clusterdeployment.NewClusterDeployment(
				clusterdeployment.WithClusterTemplate(testTemplateName),
				clusterdeployment.WithCredential(testCredentialName),
				clusterdeployment.WithAdoptedCluster(&kcmv1.AdoptedClusterConfig{
					Enabled:       true,
					DiscoveryMode: "auto",
					InfrastructureMapping: &kcmv1.InfrastructureMapping{
						Provider: "invalid-provider",
						Region:   "us-west-2",
					},
				}),
			),
			existingObjects: []runtime.Object{
				mgmt,
				cred,
				providerInterface,
				template.NewClusterTemplate(
					template.WithName(testTemplateName),
					template.WithProvidersStatus(
						"infrastructure-aws",
						"control-plane-k0smotron",
						"bootstrap-k0smotron",
					),
					template.WithValidationStatus(kcmv1.TemplateValidationStatus{Valid: true}),
				),
			},
			err: "unsupported provider: invalid-provider. Supported providers: aws, azure, gcp, vsphere",
		},
		{
			name: "should fail with scaling enabled but no scaling configuration",
			ClusterDeployment: clusterdeployment.NewClusterDeployment(
				clusterdeployment.WithClusterTemplate(testTemplateName),
				clusterdeployment.WithCredential(testCredentialName),
				clusterdeployment.WithAdoptedCluster(&kcmv1.AdoptedClusterConfig{
					Enabled: true,
					LifecycleManagement: &kcmv1.LifecycleManagement{
						Enabled: true,
						Scaling: &kcmv1.ScalingConfig{
							Enabled: true,
							// Missing MinNodes, MaxNodes, etc.
						},
					},
					InfrastructureMapping: &kcmv1.InfrastructureMapping{
						Provider: "aws",
						Region:   "us-west-2",
					},
				}),
			),
			existingObjects: []runtime.Object{
				mgmt,
				cred,
				providerInterface,
				template.NewClusterTemplate(
					template.WithName(testTemplateName),
					template.WithProvidersStatus(
						"infrastructure-aws",
						"control-plane-k0smotron",
						"bootstrap-k0smotron",
					),
					template.WithValidationStatus(kcmv1.TemplateValidationStatus{Valid: true}),
				),
			},
			err: "scaling configuration requires MinNodes and MaxNodes",
		},
		{
			name: "should fail with invalid scaling configuration",
			ClusterDeployment: clusterdeployment.NewClusterDeployment(
				clusterdeployment.WithClusterTemplate(testTemplateName),
				clusterdeployment.WithCredential(testCredentialName),
				clusterdeployment.WithAdoptedCluster(&kcmv1.AdoptedClusterConfig{
					Enabled: true,
					LifecycleManagement: &kcmv1.LifecycleManagement{
						Enabled: true,
						Scaling: &kcmv1.ScalingConfig{
							Enabled:  true,
							MinNodes: 5,
							MaxNodes: 3, // MaxNodes < MinNodes
						},
					},
					InfrastructureMapping: &kcmv1.InfrastructureMapping{
						Provider: "aws",
						Region:   "us-west-2",
					},
				}),
			),
			existingObjects: []runtime.Object{
				mgmt,
				cred,
				providerInterface,
				template.NewClusterTemplate(
					template.WithName(testTemplateName),
					template.WithProvidersStatus(
						"infrastructure-aws",
						"control-plane-k0smotron",
						"bootstrap-k0smotron",
					),
					template.WithValidationStatus(kcmv1.TemplateValidationStatus{Valid: true}),
				),
			},
			err: "MaxNodes must be greater than or equal to MinNodes",
		},
		{
			name: "should fail with invalid upgrade configuration",
			ClusterDeployment: clusterdeployment.NewClusterDeployment(
				clusterdeployment.WithClusterTemplate(testTemplateName),
				clusterdeployment.WithCredential(testCredentialName),
				clusterdeployment.WithAdoptedCluster(&kcmv1.AdoptedClusterConfig{
					Enabled: true,
					LifecycleManagement: &kcmv1.LifecycleManagement{
						Enabled: true,
						Upgrades: &kcmv1.UpgradeConfig{
							Enabled:         true,
							UpgradeStrategy: "invalid-strategy",
						},
					},
					InfrastructureMapping: &kcmv1.InfrastructureMapping{
						Provider: "aws",
						Region:   "us-west-2",
					},
				}),
			),
			existingObjects: []runtime.Object{
				mgmt,
				cred,
				providerInterface,
				template.NewClusterTemplate(
					template.WithName(testTemplateName),
					template.WithProvidersStatus(
						"infrastructure-aws",
						"control-plane-k0smotron",
						"bootstrap-k0smotron",
					),
					template.WithValidationStatus(kcmv1.TemplateValidationStatus{Valid: true}),
				),
			},
			err: "invalid upgrade strategy: invalid-strategy. Must be 'rolling' or 'in-place'",
		},
		{
			name: "should fail with invalid backup configuration",
			ClusterDeployment: clusterdeployment.NewClusterDeployment(
				clusterdeployment.WithClusterTemplate(testTemplateName),
				clusterdeployment.WithCredential(testCredentialName),
				clusterdeployment.WithAdoptedCluster(&kcmv1.AdoptedClusterConfig{
					Enabled: true,
					LifecycleManagement: &kcmv1.LifecycleManagement{
						Enabled: true,
						Backup: &kcmv1.BackupConfig{
							Enabled:  true,
							Schedule: "invalid-cron-schedule",
						},
					},
					InfrastructureMapping: &kcmv1.InfrastructureMapping{
						Provider: "aws",
						Region:   "us-west-2",
					},
				}),
			),
			existingObjects: []runtime.Object{
				mgmt,
				cred,
				providerInterface,
				template.NewClusterTemplate(
					template.WithName(testTemplateName),
					template.WithProvidersStatus(
						"infrastructure-aws",
						"control-plane-k0smotron",
						"bootstrap-k0smotron",
					),
					template.WithValidationStatus(kcmv1.TemplateValidationStatus{Valid: true}),
				),
			},
			err: "invalid backup schedule format: invalid-cron-schedule",
		},
		{
			name: "should pass with valid manual discovery configuration",
			ClusterDeployment: clusterdeployment.NewClusterDeployment(
				clusterdeployment.WithClusterTemplate(testTemplateName),
				clusterdeployment.WithCredential(testCredentialName),
				clusterdeployment.WithAdoptedCluster(&kcmv1.AdoptedClusterConfig{
					Enabled:       true,
					DiscoveryMode: "manual",
					InfrastructureMapping: &kcmv1.InfrastructureMapping{
						Provider: "aws",
						Region:   "us-west-2",
						NodeMapping: []kcmv1.NodeMapping{
							{
								NodeName:    "worker-node-1",
								MachineName: "machine-1",
							},
							{
								NodeName:    "worker-node-2",
								MachineName: "machine-2",
							},
						},
					},
				}),
			),
			existingObjects: []runtime.Object{
				mgmt,
				cred,
				providerInterface,
				template.NewClusterTemplate(
					template.WithName(testTemplateName),
					template.WithProvidersStatus(
						"infrastructure-aws",
						"control-plane-k0smotron",
						"bootstrap-k0smotron",
					),
					template.WithValidationStatus(kcmv1.TemplateValidationStatus{Valid: true}),
				),
			},
		},
		{
			name: "should pass with valid full lifecycle configuration",
			ClusterDeployment: clusterdeployment.NewClusterDeployment(
				clusterdeployment.WithClusterTemplate(testTemplateName),
				clusterdeployment.WithCredential(testCredentialName),
				clusterdeployment.WithAdoptedCluster(&kcmv1.AdoptedClusterConfig{
					Enabled:       true,
					DiscoveryMode: "auto",
					InfrastructureMapping: &kcmv1.InfrastructureMapping{
						Provider: "aws",
						Region:   "us-west-2",
						Credentials: &kcmv1.CloudCredentials{
							SecretName: "aws-credentials",
							Namespace:  testNamespace,
						},
					},
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
						Upgrades: &kcmv1.UpgradeConfig{
							Enabled:         true,
							AutoUpgrade:     true,
							UpgradeStrategy: "rolling",
							MaxUnavailable:  1,
						},
						Monitoring: &kcmv1.MonitoringConfig{
							Enabled: true,
							HealthChecks: &kcmv1.HealthCheckConfig{
								Enabled:  true,
								Interval: "30s",
								Timeout:  "10s",
							},
						},
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
				}),
			),
			existingObjects: []runtime.Object{
				mgmt,
				cred,
				providerInterface,
				template.NewClusterTemplate(
					template.WithName(testTemplateName),
					template.WithProvidersStatus(
						"infrastructure-aws",
						"control-plane-k0smotron",
						"bootstrap-k0smotron",
					),
					template.WithValidationStatus(kcmv1.TemplateValidationStatus{Valid: true}),
				),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			client := fake.NewClientBuilder().
				WithScheme(scheme.Scheme).
				WithRuntimeObjects(tt.existingObjects...).
				Build()

			validator := &ClusterDeploymentValidator{
				Client: client,
			}

			warn, err := validator.ValidateCreate(ctx, tt.ClusterDeployment)
			if tt.err != "" {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(tt.err))
			} else {
				g.Expect(err).To(Succeed())
			}

			g.Expect(warn).To(Equal(tt.warnings))
		})
	}
}

func TestAdoptedClusterValidateUpdate(t *testing.T) {
	ctx := admission.NewContextWithRequest(t.Context(), admission.Request{
		AdmissionRequest: admissionv1.AdmissionRequest{
			Operation: admissionv1.Update,
		},
	})

	oldClusterDeployment := clusterdeployment.NewClusterDeployment(
		clusterdeployment.WithClusterTemplate(testTemplateName),
		clusterdeployment.WithCredential(testCredentialName),
		clusterdeployment.WithAdoptedCluster(&kcmv1.AdoptedClusterConfig{
			Enabled:       true,
			DiscoveryMode: "auto",
			InfrastructureMapping: &kcmv1.InfrastructureMapping{
				Provider: "aws",
				Region:   "us-west-2",
			},
		}),
	)

	tests := []struct {
		name                 string
		oldClusterDeployment *kcmv1.ClusterDeployment
		newClusterDeployment *kcmv1.ClusterDeployment
		existingObjects      []runtime.Object
		err                  string
		warnings             admission.Warnings
	}{
		{
			name:                 "should pass with valid adopted cluster configuration update",
			oldClusterDeployment: oldClusterDeployment,
			newClusterDeployment: clusterdeployment.NewClusterDeployment(
				clusterdeployment.WithClusterTemplate(testTemplateName),
				clusterdeployment.WithCredential(testCredentialName),
				clusterdeployment.WithAdoptedCluster(&kcmv1.AdoptedClusterConfig{
					Enabled:       true,
					DiscoveryMode: "manual",
					InfrastructureMapping: &kcmv1.InfrastructureMapping{
						Provider: "aws",
						Region:   "us-west-2",
						NodeMapping: []kcmv1.NodeMapping{
							{
								NodeName:    "worker-node-1",
								MachineName: "machine-1",
							},
						},
					},
				}),
			),
			existingObjects: []runtime.Object{
				mgmt,
				cred,
				providerInterface,
				template.NewClusterTemplate(
					template.WithName(testTemplateName),
					template.WithProvidersStatus(
						"infrastructure-aws",
						"control-plane-k0smotron",
						"bootstrap-k0smotron",
					),
					template.WithValidationStatus(kcmv1.TemplateValidationStatus{Valid: true}),
				),
			},
		},
		{
			name:                 "should fail when changing from auto to manual discovery without node mapping",
			oldClusterDeployment: oldClusterDeployment,
			newClusterDeployment: clusterdeployment.NewClusterDeployment(
				clusterdeployment.WithClusterTemplate(testTemplateName),
				clusterdeployment.WithCredential(testCredentialName),
				clusterdeployment.WithAdoptedCluster(&kcmv1.AdoptedClusterConfig{
					Enabled:       true,
					DiscoveryMode: "manual",
					InfrastructureMapping: &kcmv1.InfrastructureMapping{
						Provider: "aws",
						Region:   "us-west-2",
						// Missing NodeMapping
					},
				}),
			),
			existingObjects: []runtime.Object{
				mgmt,
				cred,
				providerInterface,
				template.NewClusterTemplate(
					template.WithName(testTemplateName),
					template.WithProvidersStatus(
						"infrastructure-aws",
						"control-plane-k0smotron",
						"bootstrap-k0smotron",
					),
					template.WithValidationStatus(kcmv1.TemplateValidationStatus{Valid: true}),
				),
			},
			err: "manual discovery mode requires node mapping configuration",
		},
		{
			name:                 "should fail when adding invalid scaling configuration",
			oldClusterDeployment: oldClusterDeployment,
			newClusterDeployment: clusterdeployment.NewClusterDeployment(
				clusterdeployment.WithClusterTemplate(testTemplateName),
				clusterdeployment.WithCredential(testCredentialName),
				clusterdeployment.WithAdoptedCluster(&kcmv1.AdoptedClusterConfig{
					Enabled:       true,
					DiscoveryMode: "auto",
					InfrastructureMapping: &kcmv1.InfrastructureMapping{
						Provider: "aws",
						Region:   "us-west-2",
					},
					LifecycleManagement: &kcmv1.LifecycleManagement{
						Enabled: true,
						Scaling: &kcmv1.ScalingConfig{
							Enabled:  true,
							MinNodes: 5,
							MaxNodes: 3, // Invalid: MaxNodes < MinNodes
						},
					},
				}),
			),
			existingObjects: []runtime.Object{
				mgmt,
				cred,
				providerInterface,
				template.NewClusterTemplate(
					template.WithName(testTemplateName),
					template.WithProvidersStatus(
						"infrastructure-aws",
						"control-plane-k0smotron",
						"bootstrap-k0smotron",
					),
					template.WithValidationStatus(kcmv1.TemplateValidationStatus{Valid: true}),
				),
			},
			err: "MaxNodes must be greater than or equal to MinNodes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			client := fake.NewClientBuilder().
				WithScheme(scheme.Scheme).
				WithRuntimeObjects(tt.existingObjects...).
				Build()

			validator := &ClusterDeploymentValidator{
				Client: client,
			}

			warn, err := validator.ValidateUpdate(ctx, tt.oldClusterDeployment, tt.newClusterDeployment)
			if tt.err != "" {
				g.Expect(err).To(HaveOccurred())
				g.Expect(err.Error()).To(ContainSubstring(tt.err))
			} else {
				g.Expect(err).To(Succeed())
			}

			g.Expect(warn).To(Equal(tt.warnings))
		})
	}
}
