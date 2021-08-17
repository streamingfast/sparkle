package entity

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestEntity struct {
	Base
	String      string           `db:"string"`
	Integer     Int              `db:"integer"`
	Float       Float            `db:"float"`
	StringPtr   *string          `db:"string_ptr,nullable"`
	FloatPtr    *Float           `db:"float_ptr,nullable" `
	IntegerPtr  *Int             `db:"integer_ptr,nullable" `
	StringArray LocalStringArray `db:"string_array,nullable"`
}

func NewTestEntity(id string) *TestEntity {
	return &TestEntity{
		Base:    NewBase(id),
		Integer: NewIntFromLiteral(0),
		Float:   NewFloatFromLiteral(0),
	}
}

func TestPOI_WriteNilValue(t *testing.T) {
	poi := NewPOI("test")
	err := poi.Write("test_entities", "0x7bef660b110023fd795d101d5d63972a82438661", nil)
	require.NoError(t, err)
	digest := hex.EncodeToString(poi.Digest)
	fmt.Println(digest)
}

func TestPOI_Write(t *testing.T) {
	stringPtr := "helloworld"
	intPtr := NewInt(new(big.Int).SetUint64(3876123))
	floatPtr := NewFloat(new(big.Float).SetFloat64(1823.231))
	entity := NewTestEntity("0xb9afd8521c76c56ed4bc12c127c75f2fa9a9f2edda1468138664d4f0c324d30b")
	entity.String = "0x7bef660b110023fd795d101d5d63972a82438661"
	entity.Integer = NewInt(new(big.Int).SetInt64(27139))
	entity.Float = NewFloat(new(big.Float).SetFloat64(3.1643))
	entity.StringPtr = &stringPtr
	entity.FloatPtr = &floatPtr
	entity.IntegerPtr = &intPtr
	entity.StringArray = append(entity.StringArray, "aa", "bb", "cc")

	poiA := NewPOI("test")
	err := poiA.Write("test_entities", "0x7bef660b110023fd795d101d5d63972a82438661", entity)
	require.NoError(t, err)
	poiA.Apply()

	poiB := NewPOI("testb")
	entity.String = "0x7bef660b110023fd795d101d5d63972a82438660"
	err = poiB.Write("test_entities", "0x7bef660b110023fd795d101d5d63972a82438661", entity)
	require.NoError(t, err)
	poiB.Apply()

	poiC := NewPOI("testb")
	entity.StringArray = append(entity.StringArray, "dd")
	err = poiC.Write("test_entities", "0x7bef660b110023fd795d101d5d63972a82438661", entity)
	require.NoError(t, err)
	poiC.Apply()

	digestA := hex.EncodeToString(poiA.Digest)
	digestB := hex.EncodeToString(poiB.Digest)
	digestC := hex.EncodeToString(poiC.Digest)
	assert.NotEqual(t, digestA, digestB)
	assert.NotEqual(t, digestA, digestC)
	assert.NotEqual(t, digestB, digestC)
}
