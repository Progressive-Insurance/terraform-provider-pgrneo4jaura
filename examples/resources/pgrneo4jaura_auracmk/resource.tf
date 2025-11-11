resource "pgrneo4jaura_auracmk" "cmk" {
    tenant_id = "<YOUR TENANT ID>"
    cloud_provider = "aws"
    instance_type = "enterprise-db"
    name = "<YOUR CMK NAME>"
    region = "us-east-1"
    key_id = "<YOUR KEY ID>"
}