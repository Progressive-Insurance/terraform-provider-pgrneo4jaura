# Terraform Provider pgrneo4jaura

## Run provider tests

```shell
export TF_ACC=1
export TF_LOG=DEBUG
export TF_LOG_PATH=tflog
export PGRNEO4J_CLIENTID=***
export PGRNEO4J_CLIENTSECERET=***
cd /path/to/terraform-provider-pgrneo4jaura
$ # test all
$ go test -timeout 99999s -v ./...

$ # test sizing data lookup
$ go test -timeout 99999s -run '^TestAccPGRNeo4jAuraSizing$' -v ./...
$ # test instance resource
$ go test -timeout 99999s -run '^TestAccPGRNeo4jInstance$' -v ./...
$ # test cmk resource
$ go test -timeout 99999s -run '^TestAccPGRNeo4jCMK$' -v ./...
$ # test project configurations data lookup
$ go test -timeout 99999s -run '^TestAccPGRNeo4jAuraProjectConfigurations$' -v ./...
```



## Build provider

Run the following command to build and deploy the provider to your workstation. 

```shell
$ make
```
