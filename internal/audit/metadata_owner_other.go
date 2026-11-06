//go:build !linux

package audit

import "os"

// Ownership collection is performed by Linux node helpers only.
func hostFileOwner(_ os.FileInfo) (uint32, uint32, bool) {
	return 0, 0, false
}
