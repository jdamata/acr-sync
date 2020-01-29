# acr-sync
Sync images/tags from one ACR to another

## Installing
You can grab a pre-compiled version of acrpush in the release tab or generate your own:
```bash
go get -u github.com/jdamata/acr-sync
```

## Usage
```bash
  acr-sync                                                       : Display usage
  acr-sync acr-eng.azurecr.io acr-prod.azurecr.io                : List images and tags that would be pulled/pushed
  acr-sync acr-eng.azurecr.io acr-prod.azurecr.io --sync         : Pull/push all images/tags from acr-eng to acr-prod
  acr-sync acr-eng.azurecr.io acr-prod.azurecr.io --sync --prune : Pull/push all images/tags from acr-eng to acr-prod and prune unused local images after complete.
```
