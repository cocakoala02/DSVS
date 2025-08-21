package libdrynxencoding

import (
	libdrynx "github.com/ldsec/drynx/lib"
	libdrynxrange "github.com/ldsec/drynx/lib/range"
	libunlynx "github.com/ldsec/unlynx/lib"
	"go.dedis.ch/kyber/v3"
)

// EncodeCount computes the frequency of query results
func EncodeCount(input []int64, pubKey kyber.Point) (*libunlynx.CipherText, []int64) {
	resultEnc, resultClear, _ := EncodeSumWithProofs(input, pubKey, nil, 0, 0)
	return resultEnc, resultClear
}

// EncodeCountWithProofs computes the frequency of query results with the proof of range
func EncodeCountWithProofs(input []int64, pubKey kyber.Point, sigs []libdrynx.PublishSignature, l int64, u int64) (*libunlynx.CipherText, []int64, []libdrynxrange.CreateProof) {
	//sum the local DP's query results
	count := int64(0)
	count = input[0]
	//encrypt the local DP's query result
	sumEncrypted, r := libunlynx.EncryptIntGetR(pubKey, count)

	if sigs == nil {
		return sumEncrypted, []int64{count}, nil
	}
	//input range validation proof
	cp := libdrynxrange.CreateProof{Sigs: sigs, U: u, L: l, Secret: count, R: r, CaPub: pubKey, Cipher: *sumEncrypted}

	return sumEncrypted, []int64{count}, []libdrynxrange.CreateProof{cp}
}

// DecodeSum computes the sum of local DP's query results
func DecodeCount(result libunlynx.CipherText, secKey kyber.Scalar) int64 {
	//decrypt the query results
	return libunlynx.DecryptIntWithNeg(secKey, result)

}
