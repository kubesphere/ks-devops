# Get started

1. Install `ks-jenkins`
2. Install CRDs which located in [config/crd/bases](config/crd/bases)
3. Run the controller manager (you might need [a sample config file](config/samples/kubesphere.yaml))

# Create Pipeline via CLI

[ks](https://github.com/linuxsuren/ks) is an official client of KubeSphere. You can create a Pipeline by it.

`ks pip create --ws simple --template java --project default --skip-check -b good`
