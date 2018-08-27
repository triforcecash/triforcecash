package sign

import (
	"crypto/rand"
	"testing"
)

func BenchmarkSign(tstng *testing.B) {
	for i := 0; i < tstng.N; i++ {
		GenSign([]byte("qwerqwerqwerqwerqwerqwerqwerqwer"), []byte("qwerqwerqwerqwerqwerqwerqwerqwer"))
	}
}

func TestKiloGen(t *testing.T) {
	for i := 0; i < 1000; i++ {
		GenSign([]byte("qwerqwerqwerqwerqwerqwerqwerqwer"), []byte("qwerqwerqwerqwerqwerqwerqwerqwer"))
	}
}

func TestKiloVer(t *testing.T) {
	b := []byte("qwerqwerqwerqwerqwerqwerqwerqwer")
	s, p := GenSign(b, b)

	for i := 0; i < 1000; i++ {
		VerSign(b, s, p)
	}
}

func TestCorrectSign(tstng *testing.T) {
	priv := make([]byte, 64)
	data := make([]byte, 64)

	rand.Read(priv)
	rand.Read(data)

	s, pb := GenSign(data, priv)

	if VerSign(data, s, pb) {
		tstng.Log("OK")
	} else {
		tstng.Error("Signature is incorrect")
	}
}

func TestIncorrectSign(tstng *testing.T) {
	priv := make([]byte, 32)
	priv1 := make([]byte, 32)
	data := make([]byte, 32)
	data1 := make([]byte, 32)

	rand.Read(priv)
	rand.Read(priv1)
	rand.Read(data)
	rand.Read(data1)

	s, pb := GenSign(data, priv)
	s1, pb1 := GenSign(data1, priv1)

	if !(VerSign(data1, s, pb) || VerSign(data, s1, pb) || VerSign(data, s, pb1) || VerSign(data1, s1, pb) || VerSign(data1, s, pb1) || VerSign(data, s1, pb1)) {
		tstng.Log("OK")
	} else {
		tstng.Error("Signature is incorrect")
	}
}

func TestTryBruteforce(tstng *testing.T) {
	priv := make([]byte, 32)
	data := make([]byte, 32)
	rand.Read(priv)
	rand.Read(data)
	s, pb := GenSign(data, priv)
	for i := 0; i < 100; i++ {

		rand.Read(priv)
		s, _ = GenSign(data, priv)
		if VerSign(data, s, pb) {
			tstng.Error("Signature cracked")
		}
	}

	s, pb = GenSign(data, priv)
	for i := 0; i < 100; i++ {
		rand.Read(priv)
		_, pb = GenSign(data, priv)
		if VerSign(data, s, pb) {
			tstng.Error("Signature cracked")
		}
	}
}
