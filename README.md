# Terraform Provider pgrneo4jaura

## Run provider tests

```shell
$ cd /path/to/terraform-provider-pgrneo4jaura
$ TF_ACC=1 TF_LOG=DEBUG go test -v ./...
$ TF_ACC=1 TF_LOG=DEBUG go test -timeout 99999s -v ./...

$ # for a specific test only (example using TestAccPGRNeo4jInstance)
$ TF_ACC=1 TF_LOG=DEBUG TF_LOG_PATH=tflog go test -timeout 99999s -run '^TestAccPGRNeo4jInstance$' -v ./...
```

## Build provider

Run the following command to build and deploy the provider to your workstation. 

```shell
$ make
```
