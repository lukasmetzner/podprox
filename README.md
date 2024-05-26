# PodProx

**Create a new Kubernetes pod for every new TCP connection and proxy the traffic**

## Example Usage
``` bash
# Store remote pod configuration in configmap
kubectl create configmap remote-manifest --from-file=examples/remote.yaml
# Deploy podprox
kubectl apply -f k8s/podprox.yaml
```

## Architecture
```mermaid
flowchart LR
    Client1;
    Client2;
    Client3;

    PodProx;

    Pod-Client1;
    Pod-Client2;
    Pod-Client3;

    Client1 --> PodProx;
    Client2 --> PodProx;
    Client3 --> PodProx;

    PodProx --> Pod-Client1;
    PodProx --> Pod-Client2;
    PodProx --> Pod-Client3;
```
