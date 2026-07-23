package pluginhelper

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
)

func SecureIndex(length int) (int, error) {
	if length <= 0 {
		return 0, errors.New("length must be positive")
	}
	limit := uint64(^uint32(0)) + 1
	if uint64(length) > limit {
		return 0, errors.New("length is too large")
	}
	bound := limit - limit%uint64(length)
	for {
		var raw [4]byte
		if _, err := rand.Read(raw[:]); err != nil {
			return 0, err
		}
		value := uint64(binary.LittleEndian.Uint32(raw[:]))
		if value < bound {
			return int(value % uint64(length)), nil
		}
	}
}
