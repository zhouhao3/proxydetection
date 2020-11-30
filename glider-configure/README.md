This document describes how to configure a proxy service using glider.

1. Write the proxy into glider.conf

2. Build Dockerfile
```
docker build . -t zz/glider
```

3. Run the glider container and output the log to a file
```
docker run --rm -p 65535:8443 zz/glider 2>&1 | tee -a proxy.log
```
