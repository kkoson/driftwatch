// Package gcp provides a DriftWatch provider that collects infrastructure
// snapshots from Google Cloud Platform resources.
//
// The provider is registered under the name "gcp" and accepts the following
// configuration keys:
//
//	"project"  – GCP project ID (required)
//	"region"   – GCP region to scope the collection (optional, defaults to all)
//
// Resources currently collected:
//   - Compute Engine VM instances (via the Compute API)
//
// Authentication is handled via Application Default Credentials (ADC).
// Set GOOGLE_APPLICATION_CREDENTIALS or run `gcloud auth application-default
// login` before starting the daemon.
package gcp
