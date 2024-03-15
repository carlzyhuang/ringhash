package ringhash

type Instance interface {
	//获取实例四元组标识
	GetInstanceKey() InstanceKey

	GetWeight() int
}

// 服务实例的唯一标识
type InstanceKey struct {
	ServiceKey
	Host string
	Port int
}

// 服务的唯一标识KEY
type ServiceKey struct {
	//命名空间
	Namespace string
	//服务名
	Service string
}
