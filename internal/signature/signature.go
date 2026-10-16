package signature

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

func Sign(secret string,timestamp int64, payload []byte) string {

	//1: prepare string to sign:  "timestamp.payload"

	content := fmt.Sprintf("%d.%s",timestamp,string(payload))

	//2: conpute HMAC-SHA256 hash using the secret.
	mac := hmac.New(sha256.New,[]byte(secret))

	mac.Write([]byte(content)) // feeding the rawbytes to the SHA-256.
	
	//return signature prefixed with "v1="
	return "v1=" + hex.EncodeToString(mac.Sum(nil))
}