package entity

import (
	"crypto/md5"
)

type POI struct {
	Base
	Digest Bytes `db:"digest" csv:"digest"`
}

// watch out the word Interface is scary here, it's entity.Interface
func NewPOI(id string, updates map[string]map[string]Interface) *POI {
	return &POI{
		Base:   NewBase(id),
		Digest: generatePOI(updates),
	}
}

func generatePOI(updates map[string]map[string]Interface) []byte {
	return nil

}

// this function must be called only before saving a block
// 1. in parallel during toCSV phase
// 2. during linear processing
func (p *POI) AggregateDigest(previousAggregation *POI) {
	p.Digest = md5.Sum(append(previousAggregation.Digest, p.Digest...))
}

// FIXME This mechanism is used to know if we do a create or an update, but we have only a single entity in that table, called ethereum/mainnet or similar... so probably 'false' is what we want.
func (_ *POI) SkipDBLookup() bool {
	return false
}
