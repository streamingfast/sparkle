// Code generated by go-enum
// DO NOT EDIT!

package indexer

import (
	"fmt"
	"strings"
)

const (
	// ModeBatch is a Mode of type Batch.
	ModeBatch Mode = iota
	// ModeLive is a Mode of type Live.
	ModeLive
)

const _ModeName = "BatchLive"

var _ModeNames = []string{
	_ModeName[0:5],
	_ModeName[5:9],
}

// ModeNames returns a list of possible string values of Mode.
func ModeNames() []string {
	tmp := make([]string, len(_ModeNames))
	copy(tmp, _ModeNames)
	return tmp
}

var _ModeMap = map[Mode]string{
	0: _ModeName[0:5],
	1: _ModeName[5:9],
}

// String implements the Stringer interface.
func (x Mode) String() string {
	if str, ok := _ModeMap[x]; ok {
		return str
	}
	return fmt.Sprintf("Mode(%d)", x)
}

var _ModeValue = map[string]Mode{
	_ModeName[0:5]: 0,
	_ModeName[5:9]: 1,
}

// ParseMode attempts to convert a string to a Mode
func ParseMode(name string) (Mode, error) {
	if x, ok := _ModeValue[name]; ok {
		return x, nil
	}
	return Mode(0), fmt.Errorf("%s is not a valid Mode, try [%s]", name, strings.Join(_ModeNames, ", "))
}

// MarshalText implements the text marshaller method
func (x Mode) MarshalText() ([]byte, error) {
	return []byte(x.String()), nil
}

// UnmarshalText implements the text unmarshaller method
func (x *Mode) UnmarshalText(text []byte) error {
	name := string(text)
	tmp, err := ParseMode(name)
	if err != nil {
		return err
	}
	*x = tmp
	return nil
}
