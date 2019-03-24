package validator

import (
	"encoding/hex"
	"fmt"

	"github.com/go-phorce/dolly/xpki/certutil"
)

// SHA256Digest return sha256 digest prefixed with sha256:
func SHA256Digest(b []byte) string {
	hash := certutil.SHA256([]byte(b))
	manifestDigest := hex.EncodeToString(hash)
	return fmt.Sprintf("sha256:%s", manifestDigest)
}
