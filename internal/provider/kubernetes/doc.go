// Package kubernetes implements a driftwatch provider that collects
// configuration snapshots from Kubernetes deployments.
//
// It uses the Client interface to query the cluster, making it straightforward
// to inject a real client (e.g. client-go) or a fake for testing.
//
// Each deployment is represented as a snapshot whose attributes include the
// namespace, replica count, container image, and any labels. Drift is detected
// when any of these values change between polling cycles.
//
// Registration:
//
//	provider.Register("kubernetes", ...) is called automatically via init().
//	For production use, construct a provider directly with New() and a
//	properly configured client-go clientset wrapper.
package kubernetes
