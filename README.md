#

## Workflow

### Initialise Project

```bash
kubebuilder init --domain siemens.com --repo=calibre.siemens.com/calibrejob
```

### Create CalibreJob API

```bash
kubebuilder create api --group calibre --version v1 --kind CalibreJob --controller --resource
```

### Install CRD in Cluster

```bash
make install
```

### Build Controller Image

```bash
make docker-build docker-push IMG=gcr.io/livewyer-ops-public/calibre-demo:v1
```

### Deploy Controller to Cluster

```bash
make deploy IMG=gcr.io/livewyer-ops-public/calibre-demo:v1
```

### Deploy CalibreJob CR to Cluster

```bash
kubectl apply -f config/samples/calibre_v1_calibrejob.yaml
```

### Uninstall Controller and CRD

```bash
make undeploy IMG=gcr.io/livewyer-ops-public/calibre-demo:v1
```
