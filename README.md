# web-app-with-managed-performance

This is a simple web application with a fixed performance:
- latency (by default 100ms)
- error rate (by default 10% with buffer 1024 requests)
- limited rates requests before health go down by sending 500 (by default 200)
- limited rate before live-check go down (by default 500)

The application is:
- written on Go ;)
- adjustable to our needs on the fly <url>http://hostname:31848?latency=100&rate=200&errors=10&saturation=500</url>
- container/k8s ready
- has the custom metrics for Prometheus according https://medium.com/@zhimin.wen/custom-prometheus-metrics-for-apps-running-in-kubernetes-498d69ada7aa -> <url>http://hostname:31848/metrics</url>

In additional in the codebase the full set of the yaml files to create the own cluster.
<br>namespace ingress-nginx - for ingress-nginx and monitoring (prometeheus and grafana)
<br>namespace ewf-space - for app.
<br>To allow the communication between namespaces required ingress and roles.
<br>Horizontal Pod Autoscale with Custom Prometheus Metrics <url>https://itnext.io/horizontal-pod-autoscale-with-custom-metrics-8cb13e9d475</url>

# rate_loader
For validation if our metrics is configured properly we using second tools
<p>first version
<br>Usage: <br><code>./rate_loader -url=http://hostname:318484/worker -rate=20</code>
<p>alternative version
<br>Usage: <br><code>./rate_loader_v2 -url=http://hostname:318484/worker -rate=20</code>
<p><p>Run <code>go get -d -v</code> to download the dependencies
