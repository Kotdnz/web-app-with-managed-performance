# web-app-with-managed-performance

This is a simple web application with a fixed performance:
- latency
- limited rates requests before health went down sending 500
- limited rate before live-check went down

The application should be:
- written on Go ;)
- adjustable to our needs on the fly
- container/k8s ready
- has the custom metrics for Prometeus according https://medium.com/@zhimin.wen/custom-prometheus-metrics-for-apps-running-in-kubernetes-498d69ada7aa

# rate_loader
For validation if our metrics is configured properly we using second tools: rate_loader.go
<p>Usage: ./rate_loader -url=http://localhost:8080/worker -rate=20
