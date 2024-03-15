package ringhash

type RingHashLoadBalancer struct {
	cfg      *Config
	hashFunc HashFuncWithSeed
}

func (r *RingHashLoadBalancer) Setup(cfg *Config) (err error) {
	r.cfg.FillDefault()

	r.hashFunc, err = GetHashFunc(r.cfg.HashFunction)
	return
}

func (r *RingHashLoadBalancer) ChooseInstance(req *Request) (instance Instance, err error) {

	selector, err := r.getOrBuildHashRing(req.Cluster.GetInstanceSetCache().GetAllInstance())
	if nil != err {
		return 
	}
	index, err := selector.Select(req)
	if nil != err {
		return 
	}
	instance = req.Cluster.Instances[index]
	return
}

func (r *RingHashLoadBalancer) getOrBuildHashRing(instSet *InstanceSet) (selector ExtendedSelector, err error) {
	selector = instSet.GetSelector(0)
	if selector != nil {
		return
	}

	//防止服务刚上线或重建hash环时，由于selector为空，大量并发请求进入创建continuum逻辑，出现OOM
	instSet.GetLock().Lock()
	defer instSet.GetLock().Unlock()

	selector = instSet.GetSelector(0)
	if selector != nil {
		return
	}

	selector, err = r.NewContinuum(instSet)
	instSet.SetSelector(selector)

	return nil, nil
}
