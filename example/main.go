package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	bpid "github.com/minhajuddin/better_public_ids"
	"github.com/vmihailenco/msgpack/v5"
)

// --- ID types ---

// OrderID uses int64 fields, typical for database-backed sequential IDs.
type OrderID struct {
	ShopID   int64
	OrderSeq int64
}

// SessionID uses a UUID to identify a browser session.
type SessionID struct {
	UUID uuid.UUID
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
	session := SessionID{UUID: uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")}
	sessStr := bpid.MustSerialize(r, session)
	fmt.Println("SessionID serialized:  ", sessStr)

	sessBack, err := bpid.Deserialize[SessionID](r, sessStr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("SessionID deserialized: UUID=%s\n\n", sessBack.UUID)

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
	fmt.Println()
	jsonCodecExample()
	fmt.Println()
	msgpackExample()
	fmt.Println()
	opensslKeyRotationExample(r)
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

// JSONCodec implements bpid.Codec using encoding/json.
type JSONCodec struct{}

func (JSONCodec) Marshal(v any) ([]byte, error)     { return json.Marshal(v) }
func (JSONCodec) Unmarshal(data []byte, v any) error { return json.Unmarshal(data, v) }

func jsonCodecExample() {
	fmt.Println("========================================")
	fmt.Println("  Custom Codec (JSON)")
	fmt.Println("========================================")
	fmt.Println()

	r := bpid.MustNewRegistry(
		bpid.WithCodec(JSONCodec{}),
		bpid.WithType[OrderID]("order"),
	)

	fmt.Println("Registry:", r.Inspect())
	fmt.Println()

	order := OrderID{ShopID: 42, OrderSeq: 1001}
	s := bpid.MustSerialize(r, order)
	fmt.Println("JSON OrderID serialized:  ", s)

	back, err := bpid.Deserialize[OrderID](r, s)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("JSON OrderID deserialized: ShopID=%d OrderSeq=%d\n", back.ShopID, back.OrderSeq)
}

func opensslKeyRotationExample(r *bpid.Registry) {
	fmt.Println("========================================")
	fmt.Println("  Production Keys with OpenSSL")
	fmt.Println("========================================")
	fmt.Println()

	// Generate 32-byte (256-bit) keys with:
	//   openssl rand -base64 32 | tr '+/' '-_' | tr -d '='
	const (
		currentKeyB64 = "JGUzV-AC0ztqE97EeYj2Is_n6gr9afFpAELEUGaotCs"
		oldKey1B64    = "_YVqS8xrQotQLz5-CKS486oFj_E4koAZX7X_vQlb3LM"
		oldKey2B64    = "7k2G4k7JARnOdYajft0gCAmQLKml_A9uiic3ZFmQb5k"
	)

	currentKey, err := base64.RawURLEncoding.DecodeString(currentKeyB64)
	if err != nil {
		log.Fatal(err)
	}
	oldKey1, err := base64.RawURLEncoding.DecodeString(oldKey1B64)
	if err != nil {
		log.Fatal(err)
	}
	oldKey2, err := base64.RawURLEncoding.DecodeString(oldKey2B64)
	if err != nil {
		log.Fatal(err)
	}

	// Phase 1: sign with the oldest key (simulates a legacy ID).
	sr2 := bpid.MustNewSignedRegistry(r, oldKey2)
	invite := InviteID{Workspace: "acme-corp", Code: "xK9mQ"}
	signedOld2, _ := bpid.SignedSerialize(sr2, invite)
	fmt.Println("Signed with oldKey2:", signedOld2[:30]+"...")

	// Phase 2: sign with the previous key.
	sr1 := bpid.MustNewSignedRegistry(r, oldKey1)
	signedOld1, _ := bpid.SignedSerialize(sr1, invite)
	fmt.Println("Signed with oldKey1:", signedOld1[:30]+"...")

	// Phase 3: rotate to currentKey, keeping both old keys for verification.
	sr := bpid.MustNewSignedRegistry(r, currentKey, bpid.WithOldKeys(oldKey1, oldKey2))
	fmt.Println()
	fmt.Println("SignedRegistry:", sr.Inspect())
	fmt.Println()

	// Both old IDs still verify.
	inv2, err := bpid.SignedDeserialize[InviteID](sr, signedOld2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("oldKey2 ID valid: Workspace=%q Code=%q\n", inv2.Workspace, inv2.Code)

	inv1, err := bpid.SignedDeserialize[InviteID](sr, signedOld1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("oldKey1 ID valid: Workspace=%q Code=%q\n", inv1.Workspace, inv1.Code)

	// New IDs are signed with the current key.
	signedNew, _ := bpid.SignedSerialize(sr, invite)
	fmt.Println("Signed with current:", signedNew[:30]+"...")

	invNew, err := bpid.SignedDeserialize[InviteID](sr, signedNew)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Current key valid:  Workspace=%q Code=%q\n", invNew.Workspace, invNew.Code)
}

func msgpackExample() {
	fmt.Println("========================================")
	fmt.Println("  Custom Codec (msgpack)")
	fmt.Println("========================================")
	fmt.Println()

	r := bpid.MustNewRegistry(
		bpid.WithCodec(bpid.NewCodec(msgpack.Marshal, msgpack.Unmarshal)),
		bpid.WithType[SessionID]("sess"),
	)

	fmt.Println("Registry:", r.Inspect())
	fmt.Println()

	session := SessionID{UUID: uuid.MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")}
	s := bpid.MustSerialize(r, session)
	fmt.Println("Msgpack SessionID serialized:  ", s)

	back, err := bpid.Deserialize[SessionID](r, s)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Msgpack SessionID deserialized: UUID=%s\n", back.UUID)
}
