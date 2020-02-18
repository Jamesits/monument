package main

import (
	"golang.org/x/crypto/openpgp"
	"time"
)

type monument struct {
	// user configurable
	ownerName string
	ownerEmail string
	totalPeople int
	minimalPeople int

	// auto generated
	totalShares int
	minimalShares int
	deadSwitchShares int
	pgpEntity *openpgp.Entity
	shamirShares []string

	createdAt time.Time
}
