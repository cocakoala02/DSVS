package libdrynxencoding

import (
	"math"

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
func Decode(ciphers []libunlynx.CipherText, secKey kyber.Scalar, operation libdrynx.Operation, fixedScale int64) []float64 {
	switch operation.NameOp {
	// case "sum":
	// 	return []float64{float64(DecodeSum(ciphers[0], secKey))}
	case "sum":
		sScaled := DecodeSum(ciphers[0], secKey)

		return []float64{float64(sScaled) / float64(fixedScale)}
	// case "mean":
	// 	return []float64{DecodeMean(ciphers, secKey)}
	case "mean":
		// result[0]=sumScaled, result[1]=count
		// 你原来的 DecodeMean() 只做 sum/count；我们改成除以 S 再除 count
		return []float64{DecodeMeanWithScale(ciphers, secKey, fixedScale)}

	case "frequencyCount":
		freqCount := DecodeFreqCount(ciphers, secKey)
		result := make([]float64, len(freqCount))
		for i := range result {
			result[i] = float64(freqCount[i])
		}
		return result
	// case "min":
	// 	return []float64{float64(DecodeMin(ciphers, operation.QueryMin, secKey))}
	// case "max":
	// 	return []float64{float64(DecodeMax(ciphers, operation.QueryMin, secKey))}
	case "min":
		vScaled := float64(DecodeMin(ciphers, operation.QueryMin, secKey))
		return []float64{float64(vScaled) / float64(fixedScale)}

	case "max":
		vScaled := float64(DecodeMax(ciphers, operation.QueryMin, secKey))
		return []float64{float64(vScaled) / float64(fixedScale)}
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

func roundTo(x float64, digits int) float64 {
	if digits <= 0 {
		return math.Round(x)
	}
	p := math.Pow(10, float64(digits))
	return math.Round(x*p) / p
}

// 仅使用“调用方传来的” floatColumns 判定是否反缩放；
// roundDigits 控制显示小数位（例：2）
func DecodeWithSQLAndFloatCols(
	ciphers []libunlynx.CipherText,
	secKey kyber.Scalar,
	operation libdrynx.Operation,
	scaleS int64,
	sql string,
	floatColumns []string,
	roundDigits int,
) []any {

	if scaleS <= 0 {
		scaleS = 1
	}
	applyScale := decideScaledBySQLOnList(sql, floatColumns)

	switch operation.NameOp {
	case "sum":
		v := float64(DecodeSum(ciphers[0], secKey))
		if applyScale {
			v /= float64(scaleS)
			return []any{roundTo(v, roundDigits)}

		}
		return []any{int64(roundTo(v, roundDigits))}

	case "mean":
		// ciphers[0] = sum, ciphers[1] = count
		res := make([]int64, len(ciphers))
		wg := libunlynx.StartParallelize(len(ciphers))
		for i, ct := range ciphers {
			go func(i int, j libunlynx.CipherText) {
				defer wg.Done()
				res[i] = libunlynx.DecryptIntWithNeg(secKey, j)
			}(i, ct)
		}
		libunlynx.EndParallelize(wg)

		sum := float64(res[0])
		if applyScale {
			sum /= float64(scaleS)
		}
		cnt := float64(res[1])
		mean := 0.0
		if cnt != 0 {
			mean = sum / cnt
		}
		return []any{roundTo(mean, roundDigits)}

	case "min":
		v := float64(DecodeMin(ciphers, operation.QueryMin, secKey))
		if applyScale {
			v /= float64(scaleS)
			return []any{roundTo(v, roundDigits)}
		}
		return []any{int64(roundTo(v, roundDigits))}

	case "max":
		v := float64(DecodeMax(ciphers, operation.QueryMin, secKey))
		if applyScale {
			v /= float64(scaleS)
			return []any{roundTo(v, roundDigits)}
		}
		return []any{int64(roundTo(v, roundDigits))}

	case "count":
		return []any{int64(DecodeSum(ciphers[0], secKey))}

	default:
		// 其他操作维持原样（如需，也可拓展按列反缩放）
		cv := libunlynx.CipherVector(ciphers)
		ints := libunlynx.DecryptIntVectorWithNeg(secKey, &cv)
		out := make([]any, len(ints))
		for i, v := range ints {
			out[i] = float64(v)
		}
		return out
	}
}
