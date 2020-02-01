package keygen

import (
	"crypto/rand"
	"crypto/sha256"
	"math/big"
)

// All variables are saved globally as pointers for easy access
var p *big.Int       // First prime
var q *big.Int       // Second prime
var n *big.Int       // Product of Primes
var totient *big.Int // totient
var e *big.Int       // Public key
var d *big.Int       // Secret Key
var cipherKey *big.Int

// Parameter  'k' is the bitlength of n
func KeyGen(k int) {

	primesAreGood := false

	// Will keep generating new primes until conditions are satisfied
	for !primesAreGood {

		// Generate k/2 bit p & q (propable) primes
		var npoint big.Int
		p, _ = rand.Prime(rand.Reader, k/2)
		q, _ = rand.Prime(rand.Reader, k/2)

		// 'n' is created by multiplying our primes
		n = npoint.Mul(p, q)

		// Our public key 'e' is simply a bigint of 3 for speed
		e = big.NewInt(3)

		// Here we sub 1 from both primes to calculate the totient (p-1) & (q-1)
		p.Sub(p, big.NewInt(1))
		q.Sub(q, big.NewInt(1))

		// The totient is calculated by multiplying (p-1)*(q-1) which is now p*q since we already subtracted 1
		totient = big.NewInt(0)
		totient.Mul(p, q)

		d = big.NewInt(0)

		// We use ModInverse with 'e' and our totient such that 'd = n^-e % totient'
		d.ModInverse(e, totient)

		// 'd' cannot be zero and our two primes must not be equal.
		if d.String() != "0" {
			if p.String() != q.String() {
				primesAreGood = true
			}
		}
	}

	// We add 1 back to our primes for good measure
	p.Add(p, big.NewInt(1))
	q.Add(q, big.NewInt(1))
}

func HashMessage(m *big.Int) *big.Int {
	hash := sha256.New()
	hash.Write(m.Bytes())
	var bigHash = big.NewInt(0)
	bigHash.SetBytes(hash.Sum(nil))

	return bigHash
}

// Eksponent e not included. Hardcoded to 3
func GetPK() string {
	return n.String()
}

func GetPKInt() *big.Int {
	return n
}

func GetSKInt() *big.Int {
	return d
}

func SignToString(String string, publicKey *big.Int, secretKey *big.Int) string {
	// Change string to big-int
	signInt := big.NewInt(0)
	signInt.SetBytes([]byte(String))

	signInt = SignMessage(signInt, secretKey, publicKey)

	output := string(signInt.Bytes())

	return output
}

func VerifyFromString(Signature string, Pk string, plaintext string) bool {
	verInt := big.NewInt(0)
	verInt.SetBytes([]byte(Signature))

	msgInt := big.NewInt(0)
	msgInt.SetBytes([]byte(plaintext))

	PkInt := big.NewInt(0)
	PkInt.SetString(Pk, 10)

	return VerifyMessage(msgInt, verInt, PkInt)
}

func SignMessage(m *big.Int, secretKey *big.Int, publicKey *big.Int) *big.Int {
	hash := HashMessage(m)

	hash = hash.Exp(hash, secretKey, publicKey)

	return hash
}

func VerifyDraw(draw *big.Int, verificationCheck *big.Int, pk *big.Int) bool {
	temp := big.NewInt(0)
	temp.SetString(draw.String(), 10)

	verifiedDraw := temp.Exp(temp, e, pk)

	hash := HashMessage(verificationCheck)

	if verifiedDraw.String() == hash.String() {
		return true
	}

	return false

}
func VerifyMessage(m *big.Int, hash *big.Int, pk *big.Int) bool {

	hashTwo := HashMessage(m)
	hash = hash.Exp(hash, e, pk)

	if hash.String() == hashTwo.String() {
		//fmt.Println("... Authenticated")
		return true
	} else {
		//fmt.Println("... Transaction has been manipulated!")
		return false
	}
}
