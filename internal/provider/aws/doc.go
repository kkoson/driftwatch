// Package aws implements the driftwatch provider interface for Amazon Web
// Services. It collects EC2 instance configuration snapshots that the drift
// detector can compare across poll intervals.
//
// # Usage
//
// Construct a provider with a real AWS SDK client wrapper:
//
//	p := aws.New(mySDKClient, "us-east-1")
//	snaps, err := p.Collect(ctx)
//
// The package registers itself under the "aws" provider type via init(),
// but the registered builder requires a real SDK client and is therefore
// only suitable for integration use. In unit tests, pass a Client
// implementation directly to aws.New.
package aws
