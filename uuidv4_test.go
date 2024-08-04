package goid

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewUUIDv4_UsesDefaultRandomSource(t *testing.T) {
	uuidv4, err := NewUUIDv4()
	if err != nil {
		t.Fatal(err)
	}
	if uuidv4[6] < 0b0100_0000 || uuidv4[6] > 0b0100_1111 {
		t.Fatal(uuidv4)
	}
	if uuidv4[8] < 0b10_000000 || uuidv4[8] > 0b10_111111 {
		t.Fatal(uuidv4)
	}
}

func TestNewUUIDv4_UsesProvidedRandomSource(t *testing.T) {
	oldSource := randomSource
	defer SetSource(oldSource)
	r := &byteReader{b: 1}
	SetSource(r)

	uuidv4, err := NewUUIDv4()
	if err != nil {
		t.Fatal(err)
	}
	if uuidv4.String() != "01010101-0101-4101-8101-010101010101" {
		t.Fatal(uuidv4)
	}
}

func TestNewUUIDv4_UsesRandomPool(t *testing.T) {
	oldSource := randomSource
	defer SetSource(oldSource)
	r := &byteReader{b: 2}
	SetSource(r)

	EnableRandomPool()
	defer DisableRandomPool()

	uuidv4, err := NewUUIDv4()
	if err != nil {
		t.Fatal(err)
	}
	if uuidv4.String() != "02020202-0202-4202-8202-020202020202" {
		t.Fatal(uuidv4)
	}
	if r.count <= 16 {
		t.Fatal()
	}
}

func TestNewUUIDv4_UsingRandomPoolRotates(t *testing.T) {
	EnableRandomPool()
	defer DisableRandomPool()

	results := map[UUIDv4]struct{}{}

	count := len(pool.buffer)/16 + 1
	for i := 0; i < count; i++ {
		uuidv4, err := NewUUIDv4()
		if err != nil {
			t.Fatal(err)
		}
		results[uuidv4] = struct{}{}
	}
	if len(results) != count {
		t.Fatal()
	}
}

func TestNewUUIDv4_ReturnsRandomPoolError(t *testing.T) {
	oldSource := randomSource
	defer SetSource(oldSource)
	errReader := fmt.Errorf("reader")
	r := &errorReader{errReader}
	SetSource(r)

	EnableRandomPool()
	defer DisableRandomPool()

	_, err := NewUUIDv4()
	if err != errReader {
		t.Fatal(err)
	}
}

func TestNewUUIDv4Reader_ReturnsReaderError(t *testing.T) {
	errReader := fmt.Errorf("reader")

	r := &errorReader{errReader}
	if _, err := NewUUIDv4Reader(r); err != errReader {
		t.Fatal(err)
	}
}

func TestFromBytes(t *testing.T) {
	type caseType struct {
		b    []byte
		err  error
		desc string
	}

	cases := []caseType{
		{nil, ErrMalformedUUIDv4, "nil"},
		{make([]byte, 15), ErrMalformedUUIDv4, "too short"},
		{make([]byte, 17), ErrMalformedUUIDv4, "too long"},
		{make([]byte, 16), ErrInvalidUUIDv4Version, "wrong version"},
	}
	for i := 0; i < 16; i++ {
		invalid := make([]byte, 16)
		invalid[6] = 0x40
		invalid[8] = byte(i)<<4 | 0x0f
		var err error
		if i <= 0x07 || i >= 0x0e {
			err = ErrInvalidUUIDv4Variant
		}
		cases = append(cases, caseType{invalid, err, fmt.Sprintf("variant test %d", i)})
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s", tc.desc), func(t *testing.T) {
			_, err := FromBytes(tc.b)
			if (err != nil) != (tc.err != nil) {
				t.Fatalf("err = %v WANT %v", err, tc.err)
			}
			if tc.err != nil {
				if err != tc.err {
					t.Fatal(err)
				}
				return
			}
		})
	}
}

func TestParseUUIDv4(t *testing.T) {
	cases := []struct {
		s   string
		err error
	}{
		{"", ErrMalformedUUIDv4},
		{"1234ghijk", ErrMalformedUUIDv4},
		{"6a0c2488264b4e2689f5bffcc3f4c5cg", ErrMalformedUUIDv4},
		{"6a0c2488-264b-4e26-89f5-bffcc3f4c5cg2", ErrMalformedUUIDv4},

		{"6a0c2488-264b-4e26-89f5-bffcc3f4c5c/", ErrMalformedUUIDv4},
		{"6a0c2488-264b-4e26-89f5-bffcc3f4c5c:", ErrMalformedUUIDv4},
		{"6a0c2488-264b-4e26-89f5-bffcc3f4c5c@", ErrMalformedUUIDv4},
		{"6a0c2488-264b-4e26-89f5-bffcc3f4c5cG", ErrMalformedUUIDv4},
		{"6a0c2488-264b-4e26-89f5-bffcc3f4c5c`", ErrMalformedUUIDv4},
		{"6a0c2488-264b-4e26-89f5-bffcc3f4c5cg", ErrMalformedUUIDv4},

		{"00010203-0405-0607-0809-0a0b0c0d0e0f", ErrInvalidUUIDv4Version},
		{"00010203-0405-3607-0809-0a0b0c0d0e0f", ErrInvalidUUIDv4Version},
		{"00010203-0405-5607-8809-0a0b0c0d0e0f", ErrInvalidUUIDv4Version},

		{"00010203-0405-4607-0809-0a0b0c0d0e0f", ErrInvalidUUIDv4Variant},
		{"00010203-0405-4607-1809-0a0b0c0d0e0f", ErrInvalidUUIDv4Variant},
		{"00010203-0405-4607-2809-0a0b0c0d0e0f", ErrInvalidUUIDv4Variant},
		{"00010203-0405-4607-3809-0a0b0c0d0e0f", ErrInvalidUUIDv4Variant},
		{"00010203-0405-4607-4809-0a0b0c0d0e0f", ErrInvalidUUIDv4Variant},
		{"00010203-0405-4607-5809-0a0b0c0d0e0f", ErrInvalidUUIDv4Variant},
		{"00010203-0405-4607-6809-0a0b0c0d0e0f", ErrInvalidUUIDv4Variant},
		{"00010203-0405-4607-7809-0a0b0c0d0e0f", ErrInvalidUUIDv4Variant},
		{"00010203-0405-4607-e809-0a0b0c0d0e0f", ErrInvalidUUIDv4Variant},
		{"00010203-0405-4607-f809-0a0b0c0d0e0f", ErrInvalidUUIDv4Variant},

		// Valid variant 1
		{"00010203-0405-4607-8009-0a0b0c0d0e0f", nil},
		{"00010203-0405-4607-9009-0a0b0c0d0e0f", nil},
		{"00010203-0405-4607-a009-0a0b0c0d0e0f", nil},
		{"00010203-0405-4607-b009-0a0b0c0d0e0f", nil},
		{"00010203-0405-4607-8009-0A0B0C0D0E0F", nil},
		{"00010203-0405-4607-9009-0A0B0C0D0E0F", nil},
		{"00010203-0405-4607-A009-0A0B0C0D0E0F", nil},
		{"00010203-0405-4607-B009-0A0B0C0D0E0F", nil},

		// Valid variant 2
		{"00010203-0405-4607-c009-0a0b0c0d0e0f", nil},
		{"00010203-0405-4607-d009-0a0b0c0d0e0f", nil},
		{"00010203-0405-4607-C009-0A0B0C0D0E0F", nil},
		{"00010203-0405-4607-D009-0A0B0C0D0E0F", nil},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%s", tc.s), func(t *testing.T) {
			result, err := ParseUUIDv4(tc.s)
			if (err != nil) != (tc.err != nil) {
				t.Fatalf("err = %v WANT %v", err, tc.err)
			}
			if tc.err != nil {
				if err != tc.err {
					t.Fatal(err)
				}
				return
			}

			if upperResult, _ := ParseUUIDv4(strings.ToUpper(tc.s)); result != upperResult {
				t.Fatal()
			}
			if strings.ToLower(tc.s) != result.String() {
				t.Fatal(tc.s, result)
			}
		})
	}
}

func TestNewUUIDv4_String(t *testing.T) {
	v := UUIDv4([16]byte{0, 1, 2, 3, 4, 5, 0x46, 7, 0xa8, 9, 10, 11, 12, 13, 14, 15})
	if v.String() != "00010203-0405-4607-a809-0a0b0c0d0e0f" {
		t.Fatal()
	}
}

type byteReader struct {
	b     byte
	count int
}

func (br *byteReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = br.b
	}
	br.count += len(p)
	return len(p), nil
}

type errorReader struct {
	err error
}

func (er *errorReader) Read(p []byte) (int, error) {
	return 0, er.err
}
