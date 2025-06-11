# OLM-based deployment to a Helm-based deployment OLM migration chart

To facilitate the migration from an OLM-based deployment to a Helm-based deployment, a dedicated migration Helm chart is introduced. It runs job that removes following resources:
- Subscription,
- CSV,
- OperatorGroup

### How to use
```bash
helm install ibm-licensing ./helm-migration --namespace ibm-licensing --take-ownership # Run migration job, that will remove OLM resources
helm upgrade ibm-licensing ./deploy/argo-cd/components/license-service/helm-cluster-scoped --namespace ibm-licensing --take-ownership # Install LS using helm charts
```