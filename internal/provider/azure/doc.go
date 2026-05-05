// Package azure implements a DriftWatch provider that collects configuration
// snapshots from Azure Virtual Machine resources.
//
// Usage:
//
//	client := /* azure SDK compute client */
//	p := azure.New(client)
//	snaps, err := p.Collect(ctx)
//
// The provider is also registered under the name "azure" in the global
// provider registry so it can be referenced from configuration files.
package azure
