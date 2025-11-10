# Manage neo4j aura instance
resource "pgrneo4jaura_aurainstance" "aura" {
  tenant_id = "<YOUR TENANT ID>"
  name = "<YOUR INSTANCE NAME>"
  type = "enterprise-db"
  version = "5"
  cloud_provider = "aws"
  region = "us-east-1"
  memory = "4GB"
  paused = false
  n4jusr = true
  customer_managed_key_id = "<OPTIONAL CMK ID>"
  vector_optimized = true
  graph_analytics_plugin = false
  secondary_count = 0
}

