data "pgrneo4jaura_aurasizing" "sizing" {
	node_count = 1000000
	relationship_count = 5000000
	instance_type = "enterprise-ds"
	algorithm_categories = ["path-finding", "community-detection"]
}