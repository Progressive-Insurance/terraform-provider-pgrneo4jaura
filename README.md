# Terraform Provider pgrneo4jaura

## Run provider tests

```shell
$ cd /path/to/terraform-provider-pgrneo4jaura
$ TF_ACC=1 TF_LOG=INFO go test -v ./...
$ TF_ACC=1 TF_LOG=INFO go test -timeout 99999s -v ./...

$ # for a specific test only (example using TestAccPGRNeo4jInstance)
$ TF_ACC=1 TF_LOG=INFO go test -timeout 99999s -run TestAccPGRNeo4jInstance -v ./..
```

## Build provider

Run the following command to build and deploy the provider to your workstation. 

```shell
$ make
```

## Test sample configuration

Navigate to the `examples` directory. 

```shell
$ cd examples
```

Run the following command to initialize the workspace and apply the sample configuration.

```shell
$ terraform init && terraform apply
```
