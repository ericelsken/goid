package goid

import (
	"fmt"
	"io"
)

const (
	dash        = '-'
	hexAlphabet = "0123456789abcdef"
)

var (
	ErrMalformedUUIDv4      = fmt.Errorf("goid: malformed uuidv4")
	ErrInvalidUUIDv4Version = fmt.Errorf("goid: invalid version")
	ErrInvalidUUIDv4Variant = fmt.Errorf("goid: invalid variant")
)

type UUIDv4 [16]byte

func NewUUIDv4() (UUIDv4, error) {
	if usingPool() {
		return newUUIDv4FromPool()
	}
	return NewUUIDv4Reader(randomSource)
}

func NewUUIDv4Reader(r io.Reader) (UUIDv4, error) {
	result := [16]byte{}
	if _, err := io.ReadFull(r, result[:]); err != nil {
		return result, err
	}
	result[6] = 0x40 | (result[6] & 0x0f)               // Set version 4
	result[8] = 0b10_000000 | (result[8] & 0b0011_1111) // Set variant 1
	return result, nil
}

func newUUIDv4FromPool() (UUIDv4, error) {
	result, err := pool.next()
	if err != nil {
		return [16]byte{}, err
	}
	result[6] = 0x40 | (result[6] & 0x0f)               // Set version 4
	result[8] = 0b10_000000 | (result[8] & 0b0011_1111) // Set variant 1
	return result, nil
}

func FromBytes(b []byte) (UUIDv4, error) {
	if len(b) != 16 {
		return [16]byte{}, ErrMalformedUUIDv4
	}
	var result UUIDv4
	copy(result[:], b[:])
	return validateVersionVariant(&result)
}

func ParseUUIDv4(s string) (UUIDv4, error) {
	var result UUIDv4
	if len(s) != 16*2+4 || s[8] != dash || s[13] != dash || s[18] != dash || s[23] != dash {
		return result, ErrMalformedUUIDv4
	}

	fail := false

	result[0] = parseHexPair(s[0], s[1], &fail)
	result[1] = parseHexPair(s[2], s[3], &fail)
	result[2] = parseHexPair(s[4], s[5], &fail)
	result[3] = parseHexPair(s[6], s[7], &fail)

	result[4] = parseHexPair(s[9], s[10], &fail)
	result[5] = parseHexPair(s[11], s[12], &fail)

	result[6] = parseHexPair(s[14], s[15], &fail)
	result[7] = parseHexPair(s[16], s[17], &fail)

	result[8] = parseHexPair(s[19], s[20], &fail)
	result[9] = parseHexPair(s[21], s[22], &fail)

	result[10] = parseHexPair(s[24], s[25], &fail)
	result[11] = parseHexPair(s[26], s[27], &fail)
	result[12] = parseHexPair(s[28], s[29], &fail)
	result[13] = parseHexPair(s[30], s[31], &fail)
	result[14] = parseHexPair(s[32], s[33], &fail)
	result[15] = parseHexPair(s[34], s[35], &fail)

	if fail {
		return result, ErrMalformedUUIDv4
	}

	return validateVersionVariant(&result)
}

func validateVersionVariant(result *UUIDv4) (UUIDv4, error) {
	if result[6]&0xf0 != 0x40 {
		return *result, ErrInvalidUUIDv4Version
	}
	if result[8]&0b11_000000 != 0x80 && result[8]&0b111_00000 != 0xc0 {
		return *result, ErrInvalidUUIDv4Variant
	}
	return *result, nil
}

func parseHexPair(high, low byte, fail *bool) byte {
	return parseHexNibble(high, fail)<<4 + parseHexNibble(low, fail)
}

func parseHexNibble(b byte, fail *bool) byte {
	if b >= '0' && b <= '9' {
		return b - '0'
	} else if b >= 'a' && b <= 'f' {
		return b - 'a' + 10
	} else if b >= 'A' && b <= 'F' {
		return b - 'A' + 10
	}

	*fail = true
	return b
}

func (uuidv4 UUIDv4) String() string {
	result := [16*2 + 4]byte{}
	result[8], result[13], result[18], result[23] = dash, dash, dash, dash

	result[0] = hexAlphabet[uuidv4[0]>>4]
	result[1] = hexAlphabet[uuidv4[0]&15]
	result[2] = hexAlphabet[uuidv4[1]>>4]
	result[3] = hexAlphabet[uuidv4[1]&15]
	result[4] = hexAlphabet[uuidv4[2]>>4]
	result[5] = hexAlphabet[uuidv4[2]&15]
	result[6] = hexAlphabet[uuidv4[3]>>4]
	result[7] = hexAlphabet[uuidv4[3]&15]

	result[9] = hexAlphabet[uuidv4[4]>>4]
	result[10] = hexAlphabet[uuidv4[4]&15]
	result[11] = hexAlphabet[uuidv4[5]>>4]
	result[12] = hexAlphabet[uuidv4[5]&15]

	result[14] = hexAlphabet[uuidv4[6]>>4]
	result[15] = hexAlphabet[uuidv4[6]&15]
	result[16] = hexAlphabet[uuidv4[7]>>4]
	result[17] = hexAlphabet[uuidv4[7]&15]

	result[19] = hexAlphabet[uuidv4[8]>>4]
	result[20] = hexAlphabet[uuidv4[8]&15]
	result[21] = hexAlphabet[uuidv4[9]>>4]
	result[22] = hexAlphabet[uuidv4[9]&15]

	result[24] = hexAlphabet[uuidv4[10]>>4]
	result[25] = hexAlphabet[uuidv4[10]&15]
	result[26] = hexAlphabet[uuidv4[11]>>4]
	result[27] = hexAlphabet[uuidv4[11]&15]
	result[28] = hexAlphabet[uuidv4[12]>>4]
	result[29] = hexAlphabet[uuidv4[12]&15]
	result[30] = hexAlphabet[uuidv4[13]>>4]
	result[31] = hexAlphabet[uuidv4[13]&15]
	result[32] = hexAlphabet[uuidv4[14]>>4]
	result[33] = hexAlphabet[uuidv4[14]&15]
	result[34] = hexAlphabet[uuidv4[15]>>4]
	result[35] = hexAlphabet[uuidv4[15]&15]

	return string(result[:])
}
