package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	bpid "github.com/minhajuddin/better_public_ids"
)

// --- ID types ---

// OrderID uses int64 fields, typical for database-backed sequential IDs.
type OrderID struct {
	ShopID   int64
	OrderSeq int64
}

// SessionID uses a UUID (stored as [16]byte) to identify a browser session.
type SessionID struct {
	UUID [16]byte
}

// InviteID uses string fields, useful for human-readable or external identifiers.
type InviteID struct {
	Workspace string
	Code      string
}

func main() {
	// Create a single registry with all three types.
	r := bpid.MustNewRegistry(
		bpid.WithType[OrderID]("order"),
		bpid.WithType[SessionID]("sess"),
		bpid.WithType[InviteID]("inv"),
	)

	fmt.Println("Registry:", r.Inspect())
	fmt.Println()

	// --- int64-based ID ---
	order := OrderID{ShopID: 42, OrderSeq: 1001}
	orderStr := bpid.MustSerialize(r, order)
	fmt.Println("OrderID serialized:  ", orderStr)

	orderBack, err := bpid.Deserialize[OrderID](r, orderStr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("OrderID deserialized: ShopID=%d OrderSeq=%d\n\n", orderBack.ShopID, orderBack.OrderSeq)

	// --- UUID-based ID ---
	session := SessionID{UUID: mustParseUUID("6ba7b810-9dad-11d1-80b4-00c04fd430c8")}
	sessStr := bpid.MustSerialize(r, session)
	fmt.Println("SessionID serialized:  ", sessStr)

	sessBack, err := bpid.Deserialize[SessionID](r, sessStr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("SessionID deserialized: UUID=%s\n\n", formatUUID(sessBack.UUID))

	// --- String-based ID ---
	invite := InviteID{Workspace: "acme-corp", Code: "xK9mQ"}
	invStr := bpid.MustSerialize(r, invite)
	fmt.Println("InviteID serialized:  ", invStr)

	invBack, err := bpid.Deserialize[InviteID](r, invStr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("InviteID deserialized: Workspace=%q Code=%q\n\n", invBack.Workspace, invBack.Code)

	// --- Prefix extraction (useful for routing) ---
	fmt.Println("--- Prefix extraction ---")
	for _, s := range []string{orderStr, sessStr, invStr} {
		prefix, _ := r.Prefix(s)
		fmt.Printf("  %s  →  prefix=%q\n", s[:20]+"...", prefix)
	}

	fmt.Println()
	signedExample(r)
}

func signedExample(r *bpid.Registry) {
	fmt.Println("========================================")
	fmt.Println("  Signed Registry")
	fmt.Println("========================================")
	fmt.Println()

	// Wrap the same registry with HMAC signing.
	signingKey := []byte("my-secret-signing-key-do-not-share")
	sr := bpid.MustNewSignedRegistry(r, signingKey)

	fmt.Println("SignedRegistry:", sr.Inspect())
	fmt.Println()

	// --- Signed serialization ---
	order := OrderID{ShopID: 42, OrderSeq: 1001}
	signedStr, err := bpid.SignedSerialize(sr, order)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Signed OrderID:  ", signedStr)

	orderBack, err := bpid.SignedDeserialize[OrderID](sr, signedStr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Deserialized:     ShopID=%d OrderSeq=%d\n\n", orderBack.ShopID, orderBack.OrderSeq)

	// --- Tamper detection ---
	fmt.Println("--- Tamper detection ---")
	tampered := signedStr[:len(signedStr)-3] + "XXX"
	_, err = bpid.SignedDeserialize[OrderID](sr, tampered)
	fmt.Printf("Tampered ID rejected: %v\n\n", err)

	// --- Key rotation ---
	fmt.Println("--- Key rotation ---")

	// Simulate signing an ID with the original key.
	invite := InviteID{Workspace: "acme-corp", Code: "xK9mQ"}
	oldSignedInv, _ := bpid.SignedSerialize(sr, invite)
	fmt.Println("Signed with old key:", oldSignedInv[:30]+"...")

	// Rotate to a new key, keeping the old key for verification.
	newKey := []byte("rotated-key-2024-much-more-secure")
	srRotated := bpid.MustNewSignedRegistry(r, newKey, bpid.WithOldKeys(signingKey))

	// Old IDs still verify.
	invBack, err := bpid.SignedDeserialize[InviteID](srRotated, oldSignedInv)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Old ID still valid:  Workspace=%q Code=%q\n", invBack.Workspace, invBack.Code)

	// New IDs are signed with the rotated key.
	newSignedInv, _ := bpid.SignedSerialize(srRotated, invite)
	fmt.Println("Signed with new key:", newSignedInv[:30]+"...")

	invBack2, err := bpid.SignedDeserialize[InviteID](srRotated, newSignedInv)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("New ID valid:        Workspace=%q Code=%q\n", invBack2.Workspace, invBack2.Code)
}

// mustParseUUID parses a standard UUID string into [16]byte.
func mustParseUUID(s string) [16]byte {
	s = strings.ReplaceAll(s, "-", "")
	b, err := hex.DecodeString(s)
	if err != nil || len(b) != 16 {
		panic("invalid UUID: " + s)
	}
	return [16]byte(b)
}

// formatUUID formats [16]byte as a standard UUID string.
func formatUUID(b [16]byte) string {
	h := hex.EncodeToString(b[:])
	return h[:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:]
}
