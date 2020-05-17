package physical

import (
	"fmt"
	"math/rand"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
)

var rndSrc = rand.NewSource(time.Now().UnixNano())
var rnd = rand.New(rndSrc)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// Borrowed from https://colobu.com/2018/09/02/generate-random-string-in-Go/
func randomString(n int) *string {
	var b = make([]byte, n)
	for i, cache, remain := n-1, rndSrc.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = rndSrc.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	var ret = string(b)
	return &ret
}

func randomInt(n int) *int {
	var ret = rnd.Intn(n)
	return &ret
}

func randomFloat() *resource.Quantity {
	var ret = resource.MustParse(fmt.Sprintf("%f", rnd.Float64()))
	return &ret
}

func randomBoolean() *bool {
	var ret = rnd.Intn(2)/2 == 0
	return &ret
}
