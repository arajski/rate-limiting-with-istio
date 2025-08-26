# This file refactors the GKE cluster configuration to address the IP pool exhaustion error.
# The secondary IP ranges for pods and services have been significantly increased to prevent
# pool exhaustion issues.

# ---------------------------------------------------------------------------------------------------------------------
# PROVIDER CONFIGURATION
# ---------------------------------------------------------------------------------------------------------------------
terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = "cloudcon-2025"
  region  = "australia-southeast1"
}

# ---------------------------------------------------------------------------------------------------------------------
# NETWORK CONFIGURATION
# Set up a VPC and subnet with sufficiently large secondary IP ranges for GKE.
# ---------------------------------------------------------------------------------------------------------------------

# Creates a new VPC network.
resource "google_compute_network" "default" {
  name                    = "cloudcon-2025-vpc"
  auto_create_subnetworks = false
  enable_ula_internal_ipv6 = true
}

# Creates a subnetwork for the GKE cluster with larger secondary ranges.
resource "google_compute_subnetwork" "default" {
  name                     = "cloudcon-2025-subnet"
  ip_cidr_range            = "10.0.0.0/16"
  region                   = "australia-southeast1"
  stack_type               = "IPV4_IPV6"
  ipv6_access_type         = "INTERNAL"
  network                  = google_compute_network.default.id

  # Increased the services range from /29 to /20
  secondary_ip_range {
    range_name    = "services-range"
    ip_cidr_range = "192.168.0.0/20" # Provides 4096 IP addresses
  }

  # Corrected: Changed the pod range to a non-overlapping block from the services range.
  secondary_ip_range {
    range_name    = "pod-ranges"
    ip_cidr_range = "192.168.16.0/20" # Now provides 4096 IP addresses in a separate range
  }
}

# ---------------------------------------------------------------------------------------------------------------------
# GKE SERVICE ACCOUNT
# A dedicated service account for the cluster nodes.
# ---------------------------------------------------------------------------------------------------------------------

resource "google_service_account" "default" {
  account_id   = "cloudcon-2025-gke-sa"
  display_name = "CloudCon2025 GKE service account"
}

# ---------------------------------------------------------------------------------------------------------------------
# GKE CLUSTER
# Creates the standard GKE cluster.
# ---------------------------------------------------------------------------------------------------------------------

resource "google_container_cluster" "primary" {
  name                     = "cloudcon-2025-cluster"
  location                 = "australia-southeast1-a"
  network                  = google_compute_network.default.name
  subnetwork               = google_compute_subnetwork.default.name
  deletion_protection      = false
  initial_node_count       = 1
  remove_default_node_pool = true # Explicitly remove the default node pool to manage our own.

  ip_allocation_policy {
    cluster_secondary_range_name  = google_compute_subnetwork.default.secondary_ip_range[1].range_name
    services_secondary_range_name = google_compute_subnetwork.default.secondary_ip_range[0].range_name
  }

  timeouts {
    create = "30m"
    update = "40m"
  }
}

# ---------------------------------------------------------------------------------------------------------------------
# GKE NODE POOL
# Creates a separate node pool with the desired 3 nodes.
# ---------------------------------------------------------------------------------------------------------------------

resource "google_container_node_pool" "primary_nodes" {
  name       = "primary-node-pool"
  location   = "australia-southeast1-a"
  cluster    = google_container_cluster.primary.name
  node_count = 3

  node_config {
    service_account = google_service_account.default.email
    oauth_scopes = [
      "https://www.googleapis.com/auth/cloud-platform"
    ]
    labels = {
      foo = "bar"
    }
    tags = ["foo", "bar"]
  }
}

# ---------------------------------------------------------------------------------------------------------------------
# OUTPUTS
# Provide outputs for easy access to cluster information after deployment.
# ---------------------------------------------------------------------------------------------------------------------

output "cluster_name" {
  description = "The name of the GKE cluster."
  value       = google_container_cluster.primary.name
}

output "cluster_endpoint" {
  description = "The endpoint of the GKE cluster."
  value       = google_container_cluster.primary.endpoint
}

