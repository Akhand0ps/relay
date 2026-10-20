package signature

import (
	"testing"
	"time"

)


func TestSignAndVerify(t *testing.T){


	secret := "whsec_test_secret_12345"
	timeStamp := time.Now().Unix()
	payload := []byte(`{
	
		"type":"order.created",
		"amount":999
	}`)

	signature := Sign(secret,timeStamp,payload)

	if !Verify(secret,timeStamp,payload,signature) {
		t.Fatal("expected valid signature to verify successfully")
	} 
 
	tamperedPayload := []byte(`
	
	{
		"type":"order.created",
		"amount":300
	}
	`)
	if Verify(secret, timeStamp, tamperedPayload, signature) {
		t.Fatal("expected tampered payload to fail verification")
	} 
	if Verify("whsec_wrong_secret", timeStamp, payload, signature) {
		t.Fatal("expected wrong secret to fail verification")
	}

	if Verify(secret, timeStamp+10, payload, signature) {
		t.Fatal("expected tampered timestamp to fail verification")
	}
}