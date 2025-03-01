# Hetzner Cloud Provider

This document describes how to use the Hetzner Cloud provider with KCM.

## Prerequisites

1. A Hetzner Cloud account
2. An API token with read/write permissions
3. SSH key uploaded to Hetzner Cloud

## Configuration

The following environment variables are required:

- `HCLOUD_TOKEN`: Your Hetzner Cloud API token
- `HCLOUD_TOKEN_SECRET`: Name of the secret to store the API token (default: "hcloud-token")
- `HCLOUD_REGION`: The Hetzner Cloud region to deploy to (e.g., "fsn1", "nbg1", "hel1")
- `SSH_KEY_NAME`: Name of the SSH key in Hetzner Cloud
- `WORKER_NODE_TYPE`: Node type for worker nodes (e.g., "cx21", "cx31")
- `NODE_IMAGE`: Image to use for nodes (e.g., "ubuntu-20.04")

## Network Configuration

The provider creates a private network with the CIDR block `10.0.0.0/16` by default. This can be customized in the cluster template.

## Load Balancer

The provider automatically creates a load balancer for the control plane when enabled in the cluster template.

## Example Usage

1. Set up credentials:
   ```bash
   export HCLOUD_TOKEN="your-token"
   export HCLOUD_TOKEN_SECRET="hcloud-token"
   export NAMESPACE="default"
   ```

2. Create a cluster:
   ```bash
   export HCLOUD_REGION="fsn1"
   export SSH_KEY_NAME="your-ssh-key"
   export WORKER_NODE_TYPE="cx21"
   export NODE_IMAGE="ubuntu-20.04"
   ```

## Limitations

- The provider currently supports Kubernetes versions compatible with Cluster API v1beta1
- Load balancer support is limited to Hetzner Cloud Load Balancer products
- Node types are limited to Hetzner Cloud server types

## Troubleshooting

Common issues and their solutions:

1. **API Token Issues**:
   - Ensure the token has sufficient permissions
   - Verify the token is correctly stored in the secret

2. **Network Issues**:
   - Check if the CIDR block conflicts with existing networks
   - Verify network policies allow required communication

3. **Load Balancer Issues**:
   - Ensure the region supports the selected load balancer type
   - Check if there are sufficient resources in the project

## Support

For issues related to the Hetzner Cloud provider, please check:
- [Cluster API Provider Hetzner Documentation](https://github.com/syself/cluster-api-provider-hetzner)
- [Hetzner Cloud Documentation](https://docs.hetzner.com) 