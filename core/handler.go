package core

import (
	"bytes"
	"log"
)

func HandleRequest(blob []byte) []byte {
	tmp := Split(blob)
	if len(tmp) != 2 {
		return nil
	}

	action := string(tmp[0])
	body := tmp[1]

	switch action {
	case "test":
		log.Println(string(body))
		return body
	case "get":
		return Get("", body)
	case "getheader":
		return GetHeaders(body)
	case "sync candidates":
		return Candidates.Encode()
	case "sync txspool":
		return Txs.Encode()
	case "sync peers":
		return Peers.Encode()
	case "sync keys":
		return Keys.Encode()
	default:
		return nil
	}
}

func GetHeaders(key []byte) []byte {
	var res [][]byte
	tmp := Listblob(Split(Get(headprfx, key))).Get(0)
	if tmp == nil {
		return nil
	}
	var header *Header
	for i := 0; i < 1000; i++ {
		header = DecodeHeader(tmp)
		res = append(res, header.Encode())
		tmp = Listblob(Split(Get(headprfx, header.Prev))).Get(0)
		if tmp == nil {
			break
		}
	}
	return Join(res)
}

func HandleHeaders(blob []byte, key []byte) *Header {
	EncodedHeaders := Split(blob)
	headers := []*Header{}
	for _, encodedheader := range EncodedHeaders {
		header := DecodeHeader(encodedheader)
		headers = append(headers, header)
	}

	if len(headers) == 0 {
		return nil
	}

	if !bytes.Equal(headers[0].Hash(), key) {
		return nil
	}
	var lh *Header
	for i, h := range headers {
		if i == 0 {
			if !bytes.Equal(h.Hash(), key) {
				return nil
			}
		} else {
			if !IsPrev(h, lh) {
				return nil
			}
		}
		lh = h
		Put("tmp-"+headprfx, h.Hash(), h.Encode())
	}

	return headers[0]
}
