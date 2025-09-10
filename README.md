# CloudCon 2025 - Implementing Custom rate limiting with Istio
This is a sample project demonstrating how to implement domain driven rate limiting with istio.
The end goal is to setup rate limiting for an imaginary company called almost.ai...because it almost works :P

## Infrastructure
Terraform for basic GKE cluster is available under `/terraform` 

## Kuberenetes manifests 
All objects deployed to cluster are available under `/manifests`. 
As a prerequisite, the project requires istio to be installed, which can be done by running:
```sh
istioctl install --set profile=default -f istio-values.yaml -y
```

## Rate limiting service
A sample application which is going to serve UI and also provide SToW rate limiting configuration is available under `/ratelimiting-service`
It can be built and uploaded to the artifact repository.

## DNS records
The configured network assumes the domain is `almost.ai`. This can be quickly done by adding two records to `/etc/hosts` with the external IP of an istio gateway.
For example:
```
[external IP] api.almost.ai
[external IP] admin.almost.ai
```

The API server, which is a simple nginx server should be then available under the api domain, while the rate limiting service UI under admin.
