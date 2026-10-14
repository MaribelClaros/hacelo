// Command genvapid generates the VAPID key pair required by Web Push.
// Run it once, locally:
//
//	go run ./cmd/genvapid
//
// Store the output as the VAPID_PUBLIC_KEY and VAPID_PRIVATE_KEY secrets.
// Keep them: the pair cannot be regenerated identically.
package main

import (
	"fmt"
	"log"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func main() {
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		log.Fatalf("generating VAPID keys: %v", err)
	}
	fmt.Println("VAPID_PUBLIC_KEY=" + publicKey)
	fmt.Println("VAPID_PRIVATE_KEY=" + privateKey)
}
