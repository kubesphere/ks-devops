Before starting the APIServer, execute the following one command to clone Swagger UI:

```bash
make swagger-ui
```

Then, start the APIServer and explore all API documentation via the Swagger UI: <http://localhost:9090/apidocs/?url=http://localhost:9090/apidocs.json>.
 

* The URL pattern is like `http://ip:port/apidocs/?url=http://ip:port/apidocs.json`

---
In kubesphere enabled DevOps, you could update service type of devops-apiserver to NodePort, and then via the Swagger UI: `http://ip:NodePort/apidocs/?url=http://ip:NodePort/apidocs.json`.
