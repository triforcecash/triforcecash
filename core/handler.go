package core

import (
	"bytes"
)

func HandleRequest(blob []byte) []byte {
	tmp := Split(blob)
	if len(tmp) != 2 {
		return nil
	}

	action := string(tmp[0])
	body := tmp[1]

	switch action {
	case "get":
		return Get("", body)
	case "get stack":
		return GetStack(body)
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

const (
	StackDepth  = 1000
	MaxStackLen = 1 << 24
)

func GetStack(root []byte) []byte {
	var res = [][]byte{}

	var head *Header
	var stacklen int

	head = GetHeaderLocal(root)
	for i := 0; i < StackDepth && stacklen < MaxStackLen; i++ {
		if head == nil {
			return Join(res)
		}
		stacklen += len(head.Encode())
		res = append(res, head.Encode())
		if head.Id == 0 {
			break
		}
		head = GetHeaderLocal(head.Prev)
	}

	head = GetHeaderLocal(root)
	for i := 0; i < StackDepth && stacklen < MaxStackLen; i++ {
		if head == nil {
			return Join(res)
		}

		blob := Get(stateprfx, head.State)
		if blob == nil {
			blob = Get("tmp-", head.State)
		}
		if blob != nil && len(blob) > 0 {
			stacklen += len(blob)
			res = append(res, blob)
		}

		if head.Id == 0 {
			break
		}
		head = GetHeaderLocal(head.Prev)
	}

	head = GetHeaderLocal(root)
	for i := 0; i < StackDepth && stacklen < MaxStackLen; i++ {
		if head == nil {
			return Join(res)
		}

		blob := Get(txsprfx, head.Txs)
		if blob == nil {
			blob = Get("tmp-", head.Txs)
		}
		if blob != nil && len(blob) > 0 {
			stacklen += len(blob)
			res = append(res, blob)
		}

		if head.Id == 0 {
			break
		}
		head = GetHeaderLocal(head.Prev)
	}
	return Join(res)
}

func GetHeaderLocal(key []byte) *Header {
	blob := Get(headprfx, key)
	if blob != nil {
		return DecodeHeader(Listblob(Split(blob)).Get(0))
	}

	blob = Get("tmp-", key)
	if blob != nil {
		return DecodeHeader(blob)
	}

	return nil
}

func HandleStack(blob []byte, key []byte) {
	Stack := Split(blob)
	if len(Stack) > 0 && !bytes.Equal(Hash(Stack[0]), key) {
		return
	}
	for _, item := range Stack {
		Put("tmp-", Hash(item), item)
	}
}
