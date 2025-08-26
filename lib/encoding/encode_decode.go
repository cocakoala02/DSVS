package libdrynxencoding

import (
	"fmt"

	libdrynx "github.com/ldsec/drynx/lib"
	libdrynxrange "github.com/ldsec/drynx/lib/range"
	libunlynx "github.com/ldsec/unlynx/lib"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/onet/v3/log"
)

// Encode takes care of computing the query result and encode it for all possible operations.
func Encode(datas []int64, pubKey kyber.Point, signatures [][]libdrynx.PublishSignature, ranges []*[]int64, operation libdrynx.Operation) ([]libunlynx.CipherText, []int64, []libdrynxrange.CreateProof) {

	clearResponse := make([]int64, 0)
	encryptedResponse := make([]libunlynx.CipherText, 0)
	createPrf := make([]libdrynxrange.CreateProof, 0)
	withProofs := len(ranges) > 0 && len(signatures) > 0

	switch operation.NameOp {
	case "sum":
		tmpEncryptedResponse := &libunlynx.CipherText{}
		tmpPrfs := make([]libdrynxrange.CreateProof, 0)
		if withProofs {
			tmpEncryptedResponse, clearResponse, tmpPrfs = EncodeSumWithProofs(datas, pubKey, signatures[0], (*ranges[0])[1], (*ranges[0])[0])
		} else {
			tmpEncryptedResponse, clearResponse = EncodeSum(datas, pubKey)
		}
		encryptedResponse = []libunlynx.CipherText{*tmpEncryptedResponse}
		createPrf = tmpPrfs
		break
	case "mean":
		if withProofs {
			encryptedResponse, clearResponse, createPrf = EncodeMeanWithProofs(datas, pubKey, signatures, ranges)
		} else {
			encryptedResponse, clearResponse = EncodeMean(datas, pubKey)
		}
		// case "mean":
		// 	if withProofs {
		// 		if len(datas) == 2 {
		// 			encryptedResponse, clearResponse, createPrf = EncodeMeanFromSumCountWithProofs(
		// 				datas[0], datas[1], pubKey, signatures, ranges)
		// 		} else {
		// 			encryptedResponse, clearResponse, createPrf = EncodeMeanWithProofs(
		// 				datas, pubKey, signatures, ranges)
		// 		}
		// 	} else {
		// 		if len(datas) == 2 {
		// 			encryptedResponse, clearResponse = EncodeMeanFromSumCount(
		// 				datas[0], datas[1], pubKey)
		// 		} else {
		// 			encryptedResponse, clearResponse = EncodeMean(datas, pubKey)
		// 		}
		// 	}
		break
	case "count":
		tmpEncryptedResponse := &libunlynx.CipherText{}
		tmpPrfs := make([]libdrynxrange.CreateProof, 0)
		if withProofs {
			tmpEncryptedResponse, clearResponse, tmpPrfs = EncodeCountWithProofs(datas, pubKey, signatures[0], (*ranges[0])[1], (*ranges[0])[0])
		} else {
			tmpEncryptedResponse, clearResponse = EncodeCount(datas, pubKey)
		}
		encryptedResponse = []libunlynx.CipherText{*tmpEncryptedResponse}
		createPrf = tmpPrfs
		break

	case "min":
		if withProofs {
			encryptedResponse, clearResponse, createPrf = EncodeMinWithProofs(datas, operation.QueryMax, operation.QueryMin, pubKey, signatures, ranges)
		} else {
			encryptedResponse, clearResponse = EncodeMin(datas, operation.QueryMax, operation.QueryMin, pubKey)
		}
		break

	case "max":
		if withProofs {
			encryptedResponse, clearResponse, createPrf = EncodeMaxWithProofs(datas, operation.QueryMax, operation.QueryMin, pubKey, signatures, ranges)
		} else {
			encryptedResponse, clearResponse = EncodeMax(datas, operation.QueryMax, operation.QueryMin, pubKey)
		}
		break
	}
	return encryptedResponse, clearResponse, createPrf
}

// Decode decodes and computes the result of a query depending on the operation
func Decode(ciphers []libunlynx.CipherText, secKey kyber.Scalar, operation libdrynx.Operation) []float64 {
	switch operation.NameOp {
	case "sum":
		return []float64{float64(DecodeSum(ciphers[0], secKey))}
	case "mean":
		return []float64{DecodeMean(ciphers, secKey)}
	case "frequencyCount":
		freqCount := DecodeFreqCount(ciphers, secKey)
		result := make([]float64, len(freqCount))
		for i := range result {
			result[i] = float64(freqCount[i])
		}
		return result
	case "min":
		return []float64{float64(DecodeMin(ciphers, operation.QueryMin, secKey))}
	case "max":
		return []float64{float64(DecodeMax(ciphers, operation.QueryMin, secKey))}
	default:
		log.Info("no such operation:", operation)
		cv := libunlynx.CipherVector(ciphers)
		temp := libunlynx.DecryptIntVectorWithNeg(secKey, &cv)
		result := make([]float64, len(temp))
		for i, v := range temp {
			result[i] = float64(v)
		}
		return result
	}
}

// EncodeForFloat encodes floating points
func EncodeForFloat(xData [][]float64, yData []int64, lrParameters libdrynx.LogisticRegressionParameters, pubKey kyber.Point,
	signatures [][]libdrynx.PublishSignature, ranges []*[]int64, operation string) ([]libunlynx.CipherText, []int64, []libdrynxrange.CreateProof, error) {

	clearResponse := make([]int64, 0)
	encryptedResponse := make([]libunlynx.CipherText, 0)
	prf := make([]libdrynxrange.CreateProof, 0)
	withProofs := len(ranges) > 0
	switch operation {
	case "logistic regression":
		var err error
		if withProofs {
			encryptedResponse, clearResponse, prf, err = EncodeLogisticRegressionWithProofs(xData, yData, lrParameters, pubKey, signatures, ranges)
		} else {
			encryptedResponse, clearResponse, err = EncodeLogisticRegression(xData, yData, lrParameters, pubKey)
		}
		if err != nil {
			return nil, nil, nil, fmt.Errorf("when encoding response: %w", err)
		}
	}
	return encryptedResponse, clearResponse, prf, nil
}
