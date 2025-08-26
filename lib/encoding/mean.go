package libdrynxencoding

import (
	"fmt"

	libdrynx "github.com/ldsec/drynx/lib"
	libdrynxrange "github.com/ldsec/drynx/lib/range"
	libunlynx "github.com/ldsec/unlynx/lib"
	"go.dedis.ch/kyber/v3"
)

// EncodeMean computes the mean of query results
func EncodeMean(input []int64, pubKey kyber.Point) ([]libunlynx.CipherText, []int64) {
	resultEnc, resultClear, _ := EncodeMeanWithProofs(input, pubKey, nil, nil)
	return resultEnc, resultClear
}

// EncodeMeanWithProofs computes the mean of query results with the proof of range
func EncodeMeanWithProofs(input []int64, pubKey kyber.Point, sigs [][]libdrynx.PublishSignature, lu []*[]int64) ([]libunlynx.CipherText, []int64, []libdrynxrange.CreateProof) {
	//sum the local DP's query results
	// sum := int64(0)
	// for _, el := range input {
	// 	sum += el
	// }
	// N := int64(len(input))
	resultClear := input

	resultEncrypted := make([]libunlynx.CipherText, len(resultClear))
	resultRandomR := make([]kyber.Scalar, len(resultClear))

	wg := libunlynx.StartParallelize(len(resultClear))
	for i, v := range resultClear {
		go func(i int, v int64) {
			defer wg.Done()
			tmp, r := libunlynx.EncryptIntGetR(pubKey, v)
			resultEncrypted[i] = *tmp
			resultRandomR[i] = r
		}(i, v)

	}
	libunlynx.EndParallelize(wg)

	if sigs == nil {
		return resultEncrypted, resultClear, nil
	}
	fmt.Println(resultClear)

	fmt.Println(len(resultClear))

	createProofs := make([]libdrynxrange.CreateProof, len(resultClear))
	wg1 := libunlynx.StartParallelize(len(resultClear))
	for i, v := range resultClear {
		go func(i int, v int64) {
			defer wg1.Done()
			//input range validation proof
			createProofs[i] = libdrynxrange.CreateProof{Sigs: libdrynxrange.ReadColumn(sigs, i), U: (*lu[i])[0], L: (*lu[i])[1], Secret: v, R: resultRandomR[i], CaPub: pubKey, Cipher: resultEncrypted[i]}
		}(i, v)
	}
	libunlynx.EndParallelize(wg1)
	return resultEncrypted, resultClear, createProofs
}

// DecodeMean computes the mean of local DP's query results
func DecodeMean(result []libunlynx.CipherText, secKey kyber.Scalar) float64 {
	//decrypt the query results
	resultsClear := make([]int64, len(result))
	wg := libunlynx.StartParallelize(len(result))
	for i, j := range result {
		go func(i int, j libunlynx.CipherText) {
			defer wg.Done()
			resultsClear[i] = libunlynx.DecryptIntWithNeg(secKey, j)
		}(i, j)

	}
	libunlynx.EndParallelize(wg)
	return float64(resultsClear[0]) / float64(resultsClear[1]) //zai zhe li zuo chu fa, fu dian shu
}

// // 已聚合输入（sum,count）：无证明
// func EncodeMeanFromSumCount(sum int64, count int64, pubKey kyber.Point) ([]libunlynx.CipherText, []int64) {
// 	encSum, _ := libunlynx.EncryptIntGetR(pubKey, sum)
// 	encCnt, _ := libunlynx.EncryptIntGetR(pubKey, count)
// 	return []libunlynx.CipherText{*encSum, *encCnt}, []int64{sum, count}
// }

// // 已聚合输入（sum,count）：带证明（如需对 sum 与 count 分别做范围证明）
// func EncodeMeanFromSumCountWithProofs(sum int64, count int64, pubKey kyber.Point,
// 	sigs [][]libdrynx.PublishSignature, ranges []*[]int64) ([]libunlynx.CipherText, []int64, []libdrynxrange.CreateProof) {

// 	encSum, r1 := libunlynx.EncryptIntGetR(pubKey, sum)
// 	encCnt, r2 := libunlynx.EncryptIntGetR(pubKey, count)

// 	prfs := make([]libdrynxrange.CreateProof, 0, 2)
// 	if len(sigs) > 0 && len(ranges) > 0 {
// 		// 这里用你们的风格：每个值对应一列签名+一个 [L,U]
// 		// 注意：若提供两组签名/区间，分别用于 sum 和 count；否则可只对 sum 做证明
// 		prfs = append(prfs, libdrynxrange.CreateProof{
// 			Sigs:   sigs[0],
// 			U:      (*ranges[0])[0], // 你原 EncodeMeanWithProofs 里用的是 [0]=U, [1]=L，注意和现有保持一致
// 			L:      (*ranges[0])[1],
// 			Secret: sum, R: r1, CaPub: pubKey, Cipher: *encSum,
// 		})
// 		if len(sigs) > 1 && len(ranges) > 1 {
// 			prfs = append(prfs, libdrynxrange.CreateProof{
// 				Sigs:   sigs[1],
// 				U:      (*ranges[1])[0],
// 				L:      (*ranges[1])[1],
// 				Secret: count, R: r2, CaPub: pubKey, Cipher: *encCnt,
// 			})
// 		}
// 	}
// 	return []libunlynx.CipherText{*encSum, *encCnt}, []int64{sum, count}, prfs
// }
