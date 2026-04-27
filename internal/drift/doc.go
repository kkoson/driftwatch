// Package drift provides drift detection for cloud infrastructure resources.
//
// A Detector wraps a snapshot.Store and compares each newly observed snapshot
// against the previously stored one for the same resource. When attributes
// differ, it emits an Alert on a buffered channel that callers can consume
// asynchronously.
//
// Typical usage:
//
//	store := snapshot.NewStore()
//	detector := drift.NewDetector(store)
//
//	go func() {
//		for alert := range detector.Alerts() {
//			log.Println(alert)
//		}
//	}()
//
//	detector.Observe(snap)
package drift
