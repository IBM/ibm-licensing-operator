resource "helm_release" "this-cluster-scoped" {
  name       = "${var.chart_name}-cluster-scoped"
  chart      = "${path.root}/../argo-cd/components/${var.component_dirname}/helm-cluster-scoped"

  namespace = var.namespace
  create_namespace = true
  // Add take_ownership = true when PR is merged -> https://github.com/hashicorp/terraform-provider-helm/pull/1614
}

resource "helm_release" "this" {
  name       = var.chart_name
  chart      = "${path.root}/../argo-cd/components/${var.component_dirname}/helm"

  namespace = var.namespace

  depends_on = [helm_release.this-cluster-scoped]
}