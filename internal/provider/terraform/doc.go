// Package terraform implements a driftwatch provider that detects configuration
// drift in infrastructure managed by Terraform.
//
// It supports two state sources:
//
//   - Local file: set "state_file" to an absolute or relative path of a
//     terraform.tfstate file produced by `terraform show -json` or a direct
//     state pull.
//
//   - Remote HTTP endpoint: set "state_url" to a URL that returns the state
//     JSON payload (e.g. a Terraform Cloud / Enterprise state API endpoint).
//
// Each Terraform resource instance is converted into a snapshot whose ID is
// "<type>.<name>" and whose attributes include all instance attributes plus
// the synthetic keys "tf_type" and "tf_name".
//
// Register the provider in your configuration:
//
//	[[providers]]
//	type        = "terraform"
//	state_file  = "/var/lib/terraform/prod.tfstate"
package terraform
