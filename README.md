# web-app-with-managed-performance

Sometimes we need a simple web application with a fixed performance like:
- latency
- limited rates requests before health went down sending 500
- limited rate before live-check went down

Of cause, the application should be:
- written on Go ;)
- adjustable to our needs on the fly
- container/k8s ready


# rate_loader
For validation if our metrics is configured properly we using second tools: rate_loader.go
Usage: ./rate_loader -url=http://localhost:8080/worker -rate=20
