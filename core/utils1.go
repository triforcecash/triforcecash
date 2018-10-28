package core

import (
	"io"
	"log"
	"strings"
)

func ErrorHandler(err error) {
	if err != nil {
		log.Println(err)
	}
}

func IP(addr string) string {
	return strings.Split(addr, ":")[0]
}


func DecodePrefixReader(r io.Reader) ([]byte, error) {
	prfx1 := make([]byte, 1)
	_, err := io.ReadFull(r,prfx1)
	if err != nil {
		return nil, err
	}
	l := prfx1[0]
	if l < 248 {

		blob := make([]byte, l)
		_, err := io.ReadFull(r,blob)
		return blob, err

	} else {

		prfx2 := make([]byte, l-247)
		_,err:=io.ReadFull(r,prfx2)
		if err!=nil{
			return nil,err
		}
		ll := BlobToUint64(prfx2)
		if ll > ReadLimit {
			return nil, io.EOF
		}
		blob := make([]byte, ll)
		_, err = io.ReadFull(r,blob)
		return blob, err
	}
}

func DecodePrefix(b []byte, p *int) []byte {
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
	} else {
		*p++
		lbl := l - 247
		if blen < *p+lbl {
			return nil
		}
		ln := int(BlobToUint64(b[*p : *p+lbl]))
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

func Join(b [][]byte) []byte {
	var blob []byte
	for _, e := range b {
		blob = append(blob, EncodePrefix(e)...)
	}
	return blob
}
func Split(b []byte) [][]byte {
	var splited [][]byte
	p := 0
	for elem := DecodePrefix(b, &p); elem != nil; elem = DecodePrefix(b, &p) {
		splited = append(splited, elem)
	}
	return splited
}

func EncodePrefix(blob []byte) []byte {
	l := uint64(len(blob))
	if l < 248 {
		prfx := []byte{byte(l)}
		return append(prfx, blob...)
	} else {
		encodedlen := Uint64ToBlob(l)
		ll := byte(len(encodedlen) + 247)
		prfx := []byte{ll}
		prfx = append(prfx, encodedlen...)
		return append(prfx, blob...)
	}
}

func BlobToUint64(blob []byte) uint64 {
	var res uint64
	l := len(blob) - 1

	for i, b := range blob {
		res += uint64(b) << uint64(8*(l-i))
	}
	return res
}

func Uint64ToBlob(x uint64) []byte {
	i := 7
	blob := make([]byte, 8)

	for {
		blob[i] = byte(x)
		x >>= 8
		if x == 0 {
			break
		}
		i--
	}
	return blob[i:]
}
