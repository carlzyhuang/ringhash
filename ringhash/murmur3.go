package ringhash

import (
	"hash"
	"sync"

	"github.com/modern-go/reflect2"
	"github.com/spaolacci/murmur3"
)

const DefaultHashFuncName = "murmur3"

var (
	murmur3HashPool = &sync.Pool{}
)

// 通过seed的算法获取hash值
func murmur3HashWithSeed(buf []byte, seed uint32) (uint64, error) {
	var pooled = seed == 0
	var hasher hash.Hash64
	if pooled {
		poolValue := murmur3HashPool.Get()
		if !reflect2.IsNil(poolValue) {
			hasher = poolValue.(hash.Hash64)
			hasher.Reset()
		}
	}
	if nil == hasher {
		hasher = murmur3.New64WithSeed(seed)
	}
	var value uint64
	var err error
	if err = WriteBuffer(hasher, buf); nil == err {
		value = hasher.Sum64()
	}
	if pooled {
		murmur3HashPool.Put(hasher)
	}
	return value, err
}

// 包初始化函数
func init() {
	RegisterHashFunc(DefaultHashFuncName, murmur3HashWithSeed)
}
