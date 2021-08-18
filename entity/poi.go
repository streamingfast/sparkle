package entity

import (
	"crypto/md5"
	"fmt"
	"hash"
	"time"

	"github.com/ugorji/go/codec"
)

type POI struct {
	Base
	Digest  Bytes `db:"digest" csv:"digest"`
	handler *codec.MsgpackHandle
	md5     hash.Hash
}

func (p *POI) TableName() string {
	return "poi2$"
}

var poiDefaultHandler = new(codec.MsgpackHandle)

func NewPOI(causalityRegion string) *POI {
	return &POI{
		Base:   NewBase(causalityRegion),
		Digest: nil,
		md5:    md5.New(),
	}
}

func (p *POI) Clear() {
	p.md5 = md5.New()
	p.Digest = nil
}

func (p *POI) IsFinal(_ uint64, _ time.Time) bool {
	return true
}

func (p *POI) Write(entityType, entityID string, entityData interface{}) error {
	var b []byte
	enc := codec.NewEncoderBytes(&b, poiDefaultHandler)
	err := enc.Encode(entityData)
	if err != nil {
		return fmt.Errorf("unable to encode entity for poi: %w", err)
	}
	if _, err := p.md5.Write([]byte(entityType)); err != nil {
		return fmt.Errorf("unable to encode entity type: %w", err)
	}
	if _, err = p.md5.Write([]byte(entityID)); err != nil {
		return fmt.Errorf("unable to encode entity ID: %w", err)
	}
	if _, err = p.md5.Write(b); err != nil {
		return fmt.Errorf("unable to encode serialized entity: %w", err)
	}
	return nil
}

func (p *POI) Apply() {
	p.Digest = p.md5.Sum(nil)
}

func (p *POI) AggregateDigest(previousDigest []byte) {
	sum := md5.New()
	_, err := sum.Write(append(previousDigest, p.Digest...))
	if err != nil {
		panic("error generating md5sum")
	}
	p.Digest = sum.Sum(nil)
}
