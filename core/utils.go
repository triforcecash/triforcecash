package core

import (
	"crypto/sha256"
	"encoding/binary"
	"golang.org/x/crypto/sha3"
)

func uint64bytes(x uint64) []byte {
	blobx := make([]byte, 8)
	binary.BigEndian.PutUint64(blobx, x)
	return blobx
}

func bytesuint64(blobx []byte) uint64 {
	return binary.BigEndian.Uint64(blobx)
}

func boolbytes(b bool) []byte {
	if b {
		return []byte{0xff}
	} else {
		return []byte{0x00}
	}
}
func bytesbool(b []byte) bool {
	if len(b) != 1 {
		return false
	}
	return b[0] == 0xff
}

func HashSha256(data []byte) []byte {
	hsh := sha256.New()
	hsh.Write(data)
	return hsh.Sum(nil)[:]
}

func Hash(data []byte) []byte {
	tmp := sha3.Sum256(data)
	return tmp[:]
}

func PutUint64(v uint64) []byte {
	b := make([]byte, 8)
	i := 7
	for {
		b[i] = byte(v)
		v >>= 8
		if v == 0 {
			break
		}
		i--

	}
	return b[i:]
}

func GetUint64(bts []byte) uint64 {
	ln := len(bts)
	v := uint64(0)
	for i, b := range bts {
		v += uint64(b) << uint64(8*(ln-i-1))
	}
	return v
}

func encodeprfx(b []byte) []byte {
	l := len(b)

	if l < 248 {

		prfx := []byte{byte(l)}

		return append(prfx, b...)
	}

	if l >= 248 {

		bl := PutUint64(uint64(l))

		lbl := len(bl)

		prfx := []byte{byte(lbl + 247)}

		prfx = append(prfx, bl...)

		return append(prfx, b...)
	}

	return nil
}

func decodeprfx(b []byte, p *int) []byte {

	blen := len(b)

	if blen <= *p {
		return nil
	}

	l := int(b[*p])

	if l < 248 {

		*p++

		if blen < l+*p {
			return nil
		}

		res := b[*p : l+*p]

		*p += l

		return res
	}

	if l >= 248 {

		*p++

		lbl := l - 247

		if blen < *p+lbl {
			return nil
		}

		ln := int(GetUint64(b[*p : *p+lbl]))

		*p += lbl

		if blen < ln+*p {
			return nil
		}

		res := b[*p : ln+*p]

		*p += ln

		return res
	}

	return nil
}

type Listblob [][]byte

func (self Listblob) Get(i int) []byte {
	if i >= len(self) {
		return nil
	}
	return self[i]
}

func (self Listblob) GetBool(i int) bool {
	return bytesbool(self.Get(i))
}

func (self Listblob) Getuint8(i int) uint8 {
	if len(self.Get(i)) == 1 {
		return uint8(self.Get(i)[0])
	} else {
		return 0
	}
}

func (self Listblob) Getuint64(i int) uint64 {
	return GetUint64(self.Get(i))
}
