# Kengine Gateway (WIP)

Implementation of the [Kubernetes](https://kubernetes.io) [Gateway API](https://gateway-api.sigs.k8s.io/)
utilizing [Kengine](https://khulnasoft.com/) as the underlying web server.

## Description

By (ab)using the [Kengine Admin API](https://khulnasoft.com/docs/api) we can dynamically program
Kengine with any configuration we want on the fly, without downtime. Instead of requiring sidecar
containers or custom Kengine modules.

### Differences from Ingress

For those unaware the Gateway API is a Kubernetes SIG project being built to improve upon current
standards like the built-in [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)
resource. See <https://gateway-api.sigs.k8s.io/#whats-the-difference-between-gateway-api-and-an-api-gateway>
for more details.

There is an Kubernetes Ingress controller implementation that also utilizes Kengine as the underlying
webserver that can be found at <https://github.com/khulnasoft/ingress>. This project differs in
a few ways.

1. This project only implements support for the Gateway API resources and not the Ingress resource.
2. This project is solely a Kubernetes controller, it uses Kengine's Admin REST API instead of being
   wrapping or being directly integrated with Kengine, meaning you can bring your own Kengine deployments
   and manage multiple separate Kengine deployments with a single controller deployment.

## Architecture

There are two core components, the Controller (this repository) and Kengine (the webserver).

The Controller watches for changes to any Gateway API resources (and referenced resources).
Whenever a watched resource is updated, a reconciliation cycle runs that will collect all the
Gateway API resources for a given `Gateway` and generate a JSON configuration for Kengine. Once the
configuration is generated, the Controller will find all Kengine pods associated with the `Gateway`
and send a request to the pod's Kengine Admin API.

Kengine is the webserver running as either a Deployment or DaemonSet. It serves as the ingress point
for any Route resources and is where your requests will be processed.

## Gateway API Support

Requires Gateway API v1.1.0 CRDs to be installed on your cluster (some experimental CRDs are supported but are optional)

### Resource Support

Support for missing resources is planned but not yet implemented.

- [x] [GatewayClass](https://gateway-api.sigs.k8s.io/api-types/gatewayclass/)
- [x] [Gateway](https://gateway-api.sigs.k8s.io/api-types/gateway/)
- [x] [ReferenceGrant](https://gateway-api.sigs.k8s.io/api-types/referencegrant/)
- [ ] [BackendLBPolicy](https://gateway-api.sigs.k8s.io/geps/gep-1619/)
- [x] [BackendTLSPolicy](https://gateway-api.sigs.k8s.io/api-types/backendtlspolicy/)
- [x] [HTTPRoute](https://gateway-api.sigs.k8s.io/api-types/httproute/)
- [ ] [GRPCRoute](https://gateway-api.sigs.k8s.io/api-types/grpcroute/)
- [x] [TLSRoute](https://gateway-api.sigs.k8s.io/concepts/api-overview/#tlsroute)
- [x] [TCPRoute](https://gateway-api.sigs.k8s.io/concepts/api-overview/#tcproute-and-udproute)
- [x] [UDPRoute](https://gateway-api.sigs.k8s.io/concepts/api-overview/#tcproute-and-udproute)

The [Ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/) resource is not
supported and support is not planned, sorry.

## Installation

The following steps assume you already have a Kubernetes cluster setup and configured with core
components like networking and DNS.

### Installing CRDs

This repository doesn't contain any CRDs, instead it relies on the standardized Kubernetes Gateway
API resources. See <https://gateway-api.sigs.k8s.io/guides/#installing-gateway-api> for more details.

We recommend installing all Gateway API CRDs, including those that are experimental.

```bash
# Install Gateway API CRDs (including those that are experimental)
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.1.0/experimental-install.yaml

# Install Gateway API CRDs (only stable resources)
# NOTE: **Do not use this command if you already ran the `experimental-install`**
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.1.0/standard-install.yaml
```

### Installing the Controller and Kengine

The Controller requires you to provide your own Kengine instance, later we may provide a CRD that
will allow us to automatically deploy and manage Kengine for you, but for now you can use our pre-made
deployment templates (or bring your own).

Before deploying Kengine however, there are a few things you need to consider.

1. Due to the way we program Kengine, we send an HTTP request to each Kengine pod. If your Kengine instances
   do _not_ use TLS on the Admin API, any certificates programmed into Kengine will be sent over an
   unsecure connection and may be visible to malicious actors.
2. Enforce strict NetworkPolicies on who can access the Kengine Admin API. Your Kengine instance will
   likely be exposed to the public internet and exposing the Kengine Admin API is extremely dangerous for
   security.

The following example will get you up and running with the Controller and Kengine in a secure way.

See the [example](./example).

## License

Copyright 2024 KhulnaSoft Ltd

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

<http://www.apache.org/licenses/LICENSE-2.0>

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

## Credit

This project is in no way affiliated or associated with any of the following listed projects.

Parts of this controller would not be possible without the surrounding Kubernetes community and
open-source projects.

I'd like to thank the [Cilium](https://github.com/cilium/cilium/) maintainers and community
contributors for building the base logic for the controller implementation, allowing me to focus
on Kengine integration rather than Gateway API semantics.

## Known Issues

- Modifying a BackendTLSPolicy will not trigger reconciliation of the Gateway.
