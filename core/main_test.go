package core

import (
	"testing"
)

func TestBlobToUint64(t *testing.T) {
	t.Log([]byte{1, 0}, BlobToUint64([]byte{1, 0}), Uint64ToBlob(BlobToUint64([]byte{1, 0})))
	t.Log([]byte{0, 1}, BlobToUint64([]byte{0, 1}), Uint64ToBlob(BlobToUint64([]byte{0, 1})))
	t.Log([]byte{1, 0, 0}, BlobToUint64([]byte{1, 0, 0}), Uint64ToBlob(BlobToUint64([]byte{1, 0, 0})))
	t.Log([]byte{1, 1, 1}, BlobToUint64([]byte{1, 1, 1}), Uint64ToBlob(BlobToUint64([]byte{1, 1, 1})))
	t.Log([]byte{0}, BlobToUint64([]byte{0}), Uint64ToBlob(BlobToUint64([]byte{0})))
}

func TestJoinSlit(t *testing.T) {
	tmp := [][]byte{
		[]byte(`blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1
			blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1
			blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1
			blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1blob1`),
		[]byte("blob2"),
	}
	t.Log(len(tmp[0]))
	t.Logf("%q, %q, %q", tmp, Join(tmp), Split(Join(tmp)))
	t.Log(len(Split(Join(tmp))[0]))
}

func TestIP(t *testing.T) {
	t.Log((&Peer{}).IP())
	t.Log((&Peer{RemoteAddr: "127.0.0.1:8080"}).IP())
}

func TestSplit(t *testing.T) {
	t.Log(DecodeHeader(nil))
}

func TestCleaner(t *testing.T) {
	RemoveTmp()
}

func TestReward(t *testing.T){
	reward:=uint64(0) 
	for i:=uint64(0) ;i<20e6;i++{
		reward+=Reward(i)
	}	

	t.Log(reward)
}

/*func TestStart(t *testing.T) {
	PortHTTP = ":8075"
	SetSeed([]byte("asd1f1"))
	Keys.AddSelf()
	Start()
}*/
