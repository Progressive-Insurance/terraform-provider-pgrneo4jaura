# Manage neo4j aura instance
resource "pgrneo4jaura_aurainstance" "aura" {
  tenant_id = "<YOUR TENANT ID>"
  name = "<YOUR INSTANCE NAME>"
  type = "enterprise-db"
  version = "5"
  cloud_provider = "aws"
  region = "us-east-1"
  paused = false
  memory = "4GB"
}
