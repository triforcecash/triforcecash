package sign

import (
	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/sha3"
)



func GenSign(data, seed []byte) ([]byte, []byte) {
	
	tmp:=sha3.Sum256(seed)
	pr:=ed25519.NewKeyFromSeed(tmp[:])
	sign:=ed25519.Sign(pr,data)
	return sign, pr.Public().(ed25519.PublicKey)
}
func VerSign(data, sign, pub []byte) bool {
	if len(sign)!=ed25519.SignatureSize||len(pub)!=ed25519.PublicKeySize{
		return false
	}
	return ed25519.Verify(pub,data,sign) 
}

func PrivateKeyFromSeed(seed []byte)[]byte{
	tmp:=sha3.Sum256(seed)
	return ed25519.NewKeyFromSeed(tmp[:])
	
}
