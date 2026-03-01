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
