package ringhash

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
)

const (
	//默认虚拟节点数
	DefaultVnodeCount = 10
)

// 一致性hash配置对象
type Config struct {
	HashFunction string `yaml:"hashFunction" json:"hashFunction"`
	VnodeCount   int    `yaml:"vnodeCount" json:"vnodeCount"`
}

// 检验一致性hash配置
func (c *Config) Verify() error {
	var errs error
	if c.VnodeCount <= 0 {
		errs = multierror.Append(errs, fmt.Errorf("ringhash.vnodeCount must be greater than 0"))
	}
	return errs
}

// 设置一致性hash默认值
func (c *Config) FillDefault() {
	if c.VnodeCount == 0 {
		c.VnodeCount = DefaultVnodeCount
	}
	if len(c.HashFunction) == 0 {
		c.HashFunction = DefaultHashFuncName
	}
}
