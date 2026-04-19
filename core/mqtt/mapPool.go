package mqtt

import "sync"

// 全局对象池实例
var objPool = NewObjectPool()

// objectPool 提供对象池管理功能
type objectPool struct {
	mapPool sync.Pool
}

// NewObjectPool 创建一个新的对象池
func NewObjectPool() *objectPool {
	return &objectPool{
		mapPool: sync.Pool{
			New: func() interface{} {
				return make(map[string]any)
			},
		},
	}
}

// GetMap 从对象池获取一个干净的map对象
func (p *objectPool) GetMap() map[string]any {
	m := p.mapPool.Get().(map[string]any)
	// 确保返回的map是干净的，避免可能的脏数据
	for k := range m {
		delete(m, k)
	}
	return m
}

// PutMap 将map对象归还到对象池
func (p *objectPool) PutMap(m map[string]any) {
	// 清空map
	for k := range m {
		delete(m, k)
	}
	p.mapPool.Put(m)
}

/*

// objPool 对象池，用于减少内存分配
type mapPool struct {
	pool chan map[string]any
}

// GetMap 从对象池获取map
func (p *mapPool) GetMap() map[string]any {
	select {
	case m := <-p.pool:
		return m
	default:
		return make(map[string]any)
	}
}

// PutMap 将map归还到对象池
func (p *mapPool) PutMap(m map[string]any) {
	// 清空map
	for k := range m {
		delete(m, k)
	}

	select {
	case p.pool <- m:
		// 成功归还到池
	default:
		// 池已满，丢弃
	}
}

// 全局对象池
var objPool = &mapPool{
	pool: make(chan map[string]any, 100),
}

*/
