# makeroom
k8s mutating webhook removing resource requests from pods which are being created intended for test environments

caBundle in deploy.yaml is base64 encoded cert.pem

server/hook certificate dn must be makeroom.default.svc (if you keep service name and default namespace)


```
git clone
go run github.com/tom-code/makeroom/keygen
go build github.com/tom-code/makeroom
docker build . -t hook.com/makeroom:1
#edit deploy.yaml - correct caBundle
```
