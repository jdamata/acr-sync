# acr-sync
Sync images/tags from one ACR to another

## Installing
You can grab a pre-compiled version of acrpush in the release tab or generate your own:
```bash
go get -u github.com/jdamata/acr-push
```

## Usage
```bash
  acrpush                                                     : Display usage
  acrpush acr-eng.azurecr.io acr-prod.azurecr.io              : List images and tags that would be pulled/pushed
  acrpush acr-eng.azurecr.io acr-prod.azurecr.io --push       : Pull/push all images/tags from acr-eng to acr-prod
```
