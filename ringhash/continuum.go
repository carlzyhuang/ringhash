package ringhash

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"

	"git.code.oa.com/polaris/polaris-go/pkg/algorithm/search"
	"git.code.oa.com/polaris/polaris-go/pkg/log"
	"git.code.oa.com/polaris/polaris-go/pkg/model"
	"git.code.oa.com/polaris/polaris-go/pkg/plugin/loadbalancer"
	"git.code.oa.com/polaris/polaris-go/plugin/loadbalancer/common"
	"github.com/modern-go/reflect2"
)

// 一致性hash环
type ContinuumSelector struct {
	SelectorBase
	allInstances []Instance
	ring         points
	hashFunc     HashFuncWithSeed
}

// 打印hash环
func (c ContinuumSelector) String() string {
	builder := &strings.Builder{}
	builder.WriteString("[")
	if len(c.ring) > 0 {
		for i, point := range c.ring {
			if i > 0 {
				builder.WriteString(", ")
			}
			builder.WriteString(fmt.Sprintf("{\"idx\": %d, \"hash\": %d, \"instIdx\": %d, \"instId\": "+
				"\"%s\", \"address\": \"%s:%d\"}",
				i, point.hashValue, point.index, c.allInstances[point.index].GetId(),
				c.allInstances[point.index].GetHost(), c.allInstances[point.index].GetPort()))
		}
	}
	builder.WriteString("]")
	return builder.String()
}

// 做rehash
func (c *ContinuumSelector) doRehash(
	lastHash uint64, hashValues map[uint64]string, iteration int) (uint64, string, error) {
	if iteration > maxRehashIteration {
		return 0, "", fmt.Errorf("rehash exceed max iteration %d", maxRehashIteration)
	}
	var err error
	var hashValue uint64
	buf := bytes.NewBuffer(make([]byte, 0, 8))
	if err = binary.Write(buf, binary.LittleEndian, lastHash); nil != err {
		return 0, "", err
	}
	if hashValue, err = c.hashFunc(buf.Bytes(), 0); nil != err {
		return 0, "", err
	}
	if key, ok := hashValues[hashValue]; ok {
		log.GetBaseLogger().Debugf("hash conflict between %s and %d", key, lastHash)
		return c.doRehash(hashValue, hashValues, iteration+1)
	}
	lastHashStr := strconv.FormatInt(int64(lastHash), 10)
	hashValues[hashValue] = lastHashStr
	return hashValue, lastHashStr, nil
}

// 选择实例下标
func (c *ContinuumSelector) Select(value interface{}) (int, *model.ReplicateNodes, error) {
	ringLen := len(c.ring)
	switch ringLen {
	case 0:
		return -1, nil, nil
	case 1:
		return c.ring[0].index, nil, nil
	default:
		criteria := value.(*loadbalancer.Criteria)
		hashValue, err := common.CalcHashValue(criteria, c.hashFunc)
		if nil != err {
			return -1, nil, err
		}
		targetIndex, nodes := c.selectByHashValue(hashValue, criteria.ReplicateInfo.Count)
		return targetIndex, nodes, nil
	}
}

// 通过hash值选择具体的节点
func (c *ContinuumSelector) selectByHashValue(hashValue uint64, replicateCount int) (int, *model.ReplicateNodes) {
	ringIndex := search.BinarySearch(c.ring, hashValue)
	targetPoint := &c.ring[ringIndex]
	targetIndex := targetPoint.index
	if replicateCount == 0 {
		return targetIndex, nil
	}
	replicateNodesValue := targetPoint.replicates.Load()
	if !reflect2.IsNil(replicateNodesValue) {
		replicateNodes := replicateNodesValue.(*model.ReplicateNodes)
		if replicateNodes.Count == replicateCount {
			//个数匹配，则直接获取缓存信息
			return targetIndex, replicateNodes
		}
	}
	replicateIndexes := make([]int, 0, replicateCount)
	ringSize := c.ring.Len()
	for i := 1; i < ringSize; i++ {
		if len(replicateIndexes) == replicateCount {
			break
		}
		replicRingIndex := (ringIndex + i) % ringSize
		replicRing := &c.ring[replicRingIndex]
		replicateIndex := replicRing.index
		if targetIndex == replicateIndex {
			continue
		}
		if containsIndex(replicateIndexes, replicateIndex) {
			continue
		}
		replicateIndexes = append(replicateIndexes, replicateIndex)
	}
	//加入缓存
	replicateNodes := &model.ReplicateNodes{
		AllInstances: c.allInstances,
		Count:        replicateCount,
		Indexes:      replicateIndexes,
	}
	targetPoint.replicates.Store(replicateNodes)
	return targetIndex, replicateNodes
}

// 查看数组是否包含索引
func containsIndex(replicateIndexes []int, replicateIndex int) bool {
	for _, idx := range replicateIndexes {
		if idx == replicateIndex {
			return true
		}
	}
	return false
}

// 插件自定义的实例选择器base
type SelectorBase struct {
	Id int32
}

func (s *SelectorBase) ID() int32 {
	return s.Id
}

// 一致性hash环的节点
type continuumPoint struct {
	//hash的主键
	hashKey string
	//hash值
	hashValue uint64
	//实例的数组下标
	index int
	//备份节点
	replicates *atomic.Value
}

// hash环数组
type points []continuumPoint

// 比较环中节点hash值
func (c points) Less(i, j int) bool { return c[i].hashValue < c[j].hashValue }

// 获取环长度
func (c points) Len() int { return len(c) }

// 交换位置
func (c points) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

// 获取数组下标的值
func (c points) GetValue(idx int) uint64 {
	return c[idx].hashValue
}

// 数组长度
func (c points) Count() int {
	return c.Len()
}

const (
	//最多做多少次rehash
	maxRehashIteration = 5
)

// 创建hash环
func (k *RingHashLoadBalancer) NewContinuum(instSet *InstanceSet) (selector *ContinuumSelector, err error) {

	return
}
