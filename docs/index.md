---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "pgrneo4jaura Provider"
subcategory: ""
description: |-
  Progressive Neo4j Aura Provider
---

# pgrneo4jaura Provider

Progressive Neo4j Aura Provider

## Example Usage

```terraform
provider "pgrneo4jaura" {
  client_id = "<YOUR CLIENT ID>"
  client_secret = "<YOUR CLIENT SECRET>"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `client_id` (String) Progressive Neo4j Aura API client id.
- `client_secret` (String, Sensitive) Progressive Neo4j Aura API client secret.