package ringhash

import (
	"fmt"
	"io"
)

var (
	//hash函数集合
	hashFuncMap = make(map[string]HashFuncWithSeed)
)

// 可接受种子的hash函数
type HashFuncWithSeed func([]byte, uint32) (uint64, error)

// 写入buffer
func WriteBuffer(writer io.Writer, buf []byte) error {
	total := len(buf)
	for total > 0 {
		length, err := writer.Write(buf)
		if nil != err {
			return err
		}
		total -= length
	}
	return nil
}

// 注册hash函数
func RegisterHashFunc(name string, hashFunc HashFuncWithSeed) {
	if _, ok := hashFuncMap[name]; ok {
		panic("hash function %s has already existed")
	}
	hashFuncMap[name] = hashFunc
}

// 获取函数函数
func GetHashFunc(name string) (HashFuncWithSeed, error) {
	hashFunc, ok := hashFuncMap[name]
	if !ok {
		return nil, fmt.Errorf("hashFunc %s not found", name)
	}
	return hashFunc, nil
}
