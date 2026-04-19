package mqtt

import (
	"testing"
)

// TestObjectPool 测试对象池功能
func TestObjectPool(t *testing.T) {
	// 测试获取 map
	m1 := objPool.GetMap()
	if m1 == nil {
		t.Fatal("Expected non-nil map")
	}
	if len(m1) != 0 {
		t.Errorf("Expected empty map, got %d entries", len(m1))
	}

	// 设置一些值
	m1["key1"] = "value1"
	m1["key2"] = 42

	// 归还到池
	objPool.PutMap(m1)

	// 再次获取，应该是干净的
	m2 := objPool.GetMap()
	if m2 == nil {
		t.Fatal("Expected non-nil map")
	}
	if len(m2) != 0 {
		t.Errorf("Expected empty map after PutMap, got %d entries", len(m2))
	}

	t.Log("Object pool test passed")
}

// TestObjectPoolMultiple 测试多个对象的获取和归还
func TestObjectPoolMultiple(t *testing.T) {
	maps := make([]map[string]any, 10)

	// 获取多个 map
	for i := 0; i < 10; i++ {
		maps[i] = objPool.GetMap()
		if maps[i] == nil {
			t.Fatalf("Expected non-nil map at index %d", i)
		}
		maps[i]["index"] = i
	}

	// 归还所有 map
	for i := 0; i < 10; i++ {
		objPool.PutMap(maps[i])
	}

	// 再次获取，都应该是干净的
	for i := 0; i < 10; i++ {
		m := objPool.GetMap()
		if len(m) != 0 {
			t.Errorf("Expected empty map at index %d, got %d entries", i, len(m))
		}
		objPool.PutMap(m)
	}

	t.Log("Multiple object pool test passed")
}
