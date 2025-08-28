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
  region  = "asia-southeast1"
}

resource "google_compute_network" "default" {
  name                    = "cloudcon-2025-vpc"
  auto_create_subnetworks = false
  enable_ula_internal_ipv6 = true
}

resource "google_compute_subnetwork" "default" {
  name                     = "cloudcon-2025-subnet"
  ip_cidr_range            = "10.0.0.0/16"
  region                   = "asia-southeast1"
  stack_type               = "IPV4_IPV6"
  ipv6_access_type         = "INTERNAL"
  network                  = google_compute_network.default.id

  secondary_ip_range {
    range_name    = "services-range"
    ip_cidr_range = "192.168.0.0/20"
  }

  secondary_ip_range {
    range_name    = "pod-ranges"
    ip_cidr_range = "192.168.16.0/20"
  }
}

resource "google_service_account" "default" {
  account_id   = "cloudcon-2025-gke-sa"
  display_name = "CloudCon2025 GKE service account"
}

resource "google_project_iam_member" "artifact_registry_reader" {
  project = "cloudcon-2025"
  role    = "roles/artifactregistry.reader"
  member  = "serviceAccount:${google_service_account.default.email}"
}

resource "google_container_cluster" "primary" {
  name                     = "cloudcon-2025-cluster"
  location                 = "asia-southeast1-b"
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

resource "google_container_node_pool" "primary_nodes" {
  name       = "primary-node-pool"
  location   = "asia-southeast1-b"
  cluster    = google_container_cluster.primary.name
  node_count = 3

  node_config {
    machine_type    = "t2a-standard-2"
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

resource "google_artifact_registry_repository" "docker_repo" {
  repository_id = "ratelimiting"

  location = "asia-southeast1"

  format = "DOCKER"

  description = "Docker image repository for my application."
}

output "cluster_name" {
  description = "The name of the GKE cluster."
  value       = google_container_cluster.primary.name
}

output "cluster_endpoint" {
  description = "The endpoint of the GKE cluster."
  value       = google_container_cluster.primary.endpoint
}

output "repository_url" {
  description = "The URL for pushing and pulling images."
  value       = "https://<YOUR_GCP_REGION>-docker.pkg.dev/<YOUR_GCP_PROJECT_ID>/${google_artifact_registry_repository.docker_repo.repository_id}"
}
