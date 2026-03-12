# Gateway API Configuration for IBM Licensing Operator

## Overview

IBM Licensing Operator uses **Kubernetes Gateway API** to expose the License Service externally. Gateway API is the modern, standardized way to manage ingress traffic in Kubernetes, replacing the older Ingress API.

## What Gateway API Does

Gateway API creates the following resources to expose IBM License Service:

1. **Gateway** - Entry point for external traffic with HTTP/HTTPS listeners
2. **HTTPRoute** - Routes traffic from Gateway to the License Service
3. **BackendTLSPolicy** - Secures communication between Gateway and Service
4. **ConfigMap** - Stores CA certificate for backend TLS validation



### Enable/Disable Gateway

Gateway is **enabled by default** 

```yaml
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
  name: instance
spec:
  gatewayEnabled: true  
```

### Gateway Options

Configure Gateway behavior using `gatewayOptions`:

```yaml
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
  name: instance
spec:
  gatewayEnabled: true
  gatewayOptions:
    gatewayClassName: ibm-licensing        # Gateway class name (default: "ibm-licensing")
    httpPort: 80                           # HTTP listener port (default: 8080)
    httpsPort: 8080                         # HTTPS listener port (default: 443)
    tlsSecretName: my-tls-cert            # TLS certificate secret (default: "ibm-license-service-cert-internal")
    annotations:                           # Custom annotations for Gateway
      custom.annotation/key: value
```

## Configuration Parameters

### `gatewayEnabled`
- **Type:** `boolean`
- **Default:** `true` 
- **Description:** Enable or disable Gateway API. When disabled, all Gateway resources are automatically cleaned up.

### `gatewayOptions.gatewayClassName`
- **Type:** `string`
- **Default:** `"ibm-licensing"`
- **Description:** Name of the GatewayClass to use. Must match an existing GatewayClass in your cluster.

### `gatewayOptions.httpPort`
- **Type:** `int32`
- **Default:** `80`
- **Range:** `1-65535`
- **Description:** Port for HTTP listener. Traffic on this port is unencrypted.

### `gatewayOptions.httpsPort`
- **Type:** `int32`
- **Default:** `8080`
- **Range:** `1-65535`
- **Description:** Port for HTTPS listener. Only created when `tlsSecretName` is set. Traffic is TLS-terminated at the Gateway.

### `gatewayOptions.tlsSecretName`
- **Type:** `string`
- **Default:** `"ibm-license-service-cert-internal"`
- **Description:** Name of the Kubernetes Secret containing TLS certificate for HTTPS listener. Secret must contain `tls.crt` and `tls.key` fields.

### `gatewayOptions.annotations`
- **Type:** `map[string]string`
- **Default:** `nil`
- **Description:** Custom annotations to add to the Gateway resource. Useful for configuring cloud provider-specific settings.

### this annotations thing is an Ingress's legacy so probably it may be redundant but I left it here for now