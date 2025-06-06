terraform {
  required_providers {
    helm = {
      source = "hashicorp/helm"
      version = "3.0.0-pre2"
    }
  }
}

provider "helm" {
  kubernetes = {
    config_path = "~/.kube/config"
  }
}
