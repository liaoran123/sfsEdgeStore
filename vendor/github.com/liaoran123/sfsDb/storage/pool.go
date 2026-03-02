package storage

import (
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

// 批处理对象池配置
var (
	// MaxBatchSize 批处理对象大小阈值（字节），超过此值的批处理对象不会被缓存
	MaxBatchSize = 1024 * 1024 // 1MB
)

// batchPool 包装sync.Pool，用于管理批处理对象
type batchPool struct {
	pool sync.Pool
}

// Get 从池中获取批处理对象
func (p *batchPool) Get() *leveldb.Batch {
	// 从池中获取对象
	obj := p.pool.Get()
	if obj == nil {
		// 安全保障：如果从池中获取到 nil，创建一个新的批处理对象
		return new(leveldb.Batch)
	}

	// 类型断言，确保获取的是 *leveldb.Batch 类型
	batch, ok := obj.(*leveldb.Batch)
	if !ok {
		// 类型不匹配，创建一个新的批处理对象
		return new(leveldb.Batch)
	}

	// 确保返回的是干净的批处理对象
	batch.Reset()
	return batch
}

// Put 将批处理对象放回池中
func (p *batchPool) Put(batch *leveldb.Batch) {
	// 检查批处理对象是否为 nil
	if batch == nil {
		// 批处理对象为 nil，直接返回
		return
	}

	// 检查批处理对象大小
	if batch.Len() > MaxBatchSize {
		// 批处理对象大小超过阈值，直接丢弃，让垃圾回收器处理这个对象
		batch.Reset()
		return
	}

	// 重置批处理对象
	batch.Reset()

	// 将对象放回池中
	p.pool.Put(batch)
}

// 批处理对象池
var LdbBatchPool = &batchPool{
	pool: sync.Pool{
		New: func() any {
			return new(leveldb.Batch)
		},
	},
}

// SetMaxBatchSize 设置批处理对象大小阈值（字节）
// size: 批处理对象大小阈值（字节）
func SetMaxBatchSize(size int) {
	if size > 0 {
		MaxBatchSize = size
	}
}

// GetMaxBatchSize 获取批处理对象大小阈值（字节）
func GetMaxBatchSize() int {
	return MaxBatchSize
}

// snapshotPool 包装sync.Pool，用于管理LevelDBStore类型的快照实例
type snapshotPool struct {
	pool sync.Pool
}

// Get 从池中获取快照对象
func (p *snapshotPool) Get() *LevelDBStore {
	snapshot := p.pool.Get().(*LevelDBStore)
	if snapshot == nil {
		// 安全保障：如果从池中获取到 nil，创建一个新的快照对象
		return new(LevelDBStore)
	}

	// 确保返回的是干净的快照对象
	// 即使在极端情况下，也能保证对象状态的一致性
	snapshot.ldb = nil
	snapshot.originalDB = nil
	snapshot.isSnapshot = false
	snapshot.opts = nil

	return snapshot
}

// Put 将快照对象放回池中
func (p *snapshotPool) Put(snapshot *LevelDBStore) {
	// 检查快照对象是否为 nil
	if snapshot == nil {
		// 快照对象为 nil，直接返回
		return
	}

	// 重置快照对象状态
	snapshot.ldb = nil
	snapshot.originalDB = nil
	snapshot.isSnapshot = false
	snapshot.opts = nil

	// 将对象放回池中
	p.pool.Put(snapshot)
}

// 快照对象池
var LdbSnapshotPool = &snapshotPool{
	pool: sync.Pool{
		New: func() any {
			return new(LevelDBStore)
		},
	},
}
