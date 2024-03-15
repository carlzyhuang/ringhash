package ringhash

import "sync"

type Request struct {
	//用户传入用于计算hash的二进制流
	HashKey []byte
	//用户传入用于计算hash的int值
	HashValue uint64
	//必选，目标cluster
	Cluster *Cluster
}

type Cluster struct {
	ClusterKey
	Instances []Instance

	//通过服务路由选择出来的集群，缓存起来
	cache *InstanceSetCache
}

// 重构建集群缓存
func (c *Cluster) GetInstanceSetCache() *InstanceSetCache {
	if nil != c.cache {
		return c.cache
	}
	c.cache = c.buildCluster()

	return c.cache
}

// 根据给定集群构建索引
func (c *Cluster) buildCluster() *InstanceSetCache {
	//TODO cache
	return nil
}

type ClusterKey struct {
	Key string
}

// 哈希环的实例缓存
type InstanceSet struct {
	serviceKey   ServiceKey
	allInstances []Instance
	//扩展的实例选择器
	selector *sync.Map
	//加锁，防止重复创建selector
	lock sync.Mutex
}

// 设置selector
func (i *InstanceSet) SetSelector(selector ExtendedSelector) {
	i.selector.Store(selector.ID(), selector)
}

// 获取selector
func (i *InstanceSet) GetSelector(id int32) ExtendedSelector {
	value, ok := i.selector.Load(id)
	if !ok {
		return nil
	}
	return value.(ExtendedSelector)
}

// 获取互斥锁，用于创建selector时使用，防止重复创建selector
func (i *InstanceSet) GetLock() sync.Locker {
	return &i.lock
}

// 可供插件自定义的实例选择器
type ExtendedSelector interface {
	//选择实例下标
	Select(criteria interface{}) (int, error)

	//对应负载均衡插件的名字
	ID() int32
}

type InstanceSetCache struct {
	//全量服务实例
	allInstances *InstanceSet
}

func (i *InstanceSetCache) GetAllInstance() *InstanceSet {
	return i.allInstances
}
