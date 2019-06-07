package common

// https://blog.kowalczyk.info/article/JyRZ/generating-good-unique-ids-in-go.html

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/chilts/sid"
	"github.com/kjk/betterguid"
	"github.com/oklog/ulid"
	"github.com/rs/xid"
	"github.com/satori/go.uuid"
	"github.com/segmentio/ksuid"
	"github.com/sony/sonyflake"
)

func GenXid() string {
	id := xid.New()
	fmt.Printf("github.com/rs/xid:           %s\n", id.String())
	return id.String()
}

func GenKsuid() string {
	id := ksuid.New()
	fmt.Printf("github.com/segmentio/ksuid:  %s\n", id.String())
	return id.String()
}

func GenBetterGUID() string {
	id := betterguid.New()
	fmt.Printf("github.com/kjk/betterguid:   %s\n", id)
	return id
}

func GenUlid() string {
	t := time.Now().UTC()
	entropy := rand.New(rand.NewSource(t.UnixNano()))
	id := ulid.MustNew(ulid.Timestamp(t), entropy)
	fmt.Printf("github.com/oklog/ulid:       %s\n", id.String())
	return id.String()
}

func GenSonyflake() string {
	flake := sonyflake.NewSonyflake(sonyflake.Settings{})
	id, err := flake.NextID()
	if err != nil {
		log.Fatalf("flake.NextID() failed with %s\n", err)
	}
	// Note: this is base16, could shorten by encoding as base62 string
	fmt.Printf("github.com/sony/sonyflake:   %x\n", id)
	return string(id)
}

func GenSid() string {
	id := sid.Id()
	fmt.Printf("github.com/chilts/sid:       %s\n", id)
	return id
}

func GenUUIDv4() string {
	id := uuid.NewV4()
	fmt.Printf("github.com/satori/go.uuid:   %s\n", id.String())
	return id.String()
}
