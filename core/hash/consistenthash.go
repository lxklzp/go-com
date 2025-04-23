package hash

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"sync"
)

/********** 一致性哈希
这是go-zero的组件：
支持动态新增、删除节点
支持虚拟节点、节点权重、哈希冲突处理
**********/

const (
	TopWeight   = 100      // 节点可以设置的最大权重值
	minReplicas = 100      // 每个节点的最小虚拟节点数
	prime       = 16777619 // 用于二次哈希计算的质数
)

// 用于占位，节省内存空间
var placeholder placeholderType

type placeholderType = struct{}

type (
	// Func 定义哈希函数类型
	Func func(data []byte) uint64

	// ConsistentHash 一致性哈希数据结构
	ConsistentHash struct {
		hashFunc Func                       // 哈希函数
		replicas int                        // 每个节点的默认虚拟节点数
		keys     []uint64                   // 哈希环，排序后的虚拟节点哈希键列表
		ring     map[uint64][]any           // 虚拟节点哈希键到真实节点的映射，[]any的作用是处理哈希冲突，即不同真实节点映射了相同的虚拟节点哈希键
		nodes    map[string]placeholderType // 真实节点字符串集合
		lock     sync.RWMutex               // 处理并发读写keys、ring、nodes的锁
	}
)

func NewConsistentHash() *ConsistentHash {
	return NewCustomConsistentHash(minReplicas, Hash)
}

func NewCustomConsistentHash(replicas int, fn Func) *ConsistentHash {
	if replicas < minReplicas {
		replicas = minReplicas
	}

	if fn == nil {
		fn = Hash
	}

	return &ConsistentHash{
		hashFunc: fn,
		replicas: replicas,
		ring:     make(map[uint64][]any),
		nodes:    make(map[string]placeholderType),
	}
}

// Add 使用默认虚拟节点数添加节点
func (h *ConsistentHash) Add(node any) {
	h.AddWithReplicas(node, h.replicas)
}

// AddWithReplicas 使用指定虚拟节点数添加节点
func (h *ConsistentHash) AddWithReplicas(node any, replicas int) {
	// 先删除，再添加的方式
	h.Remove(node)

	// 指定虚拟节点数应当小于默认节点
	if replicas > h.replicas {
		replicas = h.replicas
	}

	nodeRepr := repr(node) // 将节点转换成字符串
	h.lock.Lock()
	defer h.lock.Unlock()
	h.addNode(nodeRepr) // 添加真实节点 将节点字符串添加到nodes

	// 添加虚拟节点
	for i := 0; i < replicas; i++ {
		hash := h.hashFunc([]byte(nodeRepr + strconv.Itoa(i))) // 使用节点字符串+序号生成虚拟节点哈希键
		h.keys = append(h.keys, hash)                          // 将虚拟节点哈希键添加到keys
		h.ring[hash] = append(h.ring[hash], node)              // 将虚拟节点哈希键到真实节点的映射添加到ring
	}

	// 哈希环排序，用于sort.Search
	sort.Slice(h.keys, func(i, j int) bool {
		return h.keys[i] < h.keys[j]
	})
}

// AddWithWeight 根据权重添加节点（权重1-100）
func (h *ConsistentHash) AddWithWeight(node any, weight int) {
	replicas := h.replicas * weight / TopWeight
	h.AddWithReplicas(node, replicas)
}

// Get 根据查找键对应的节点
func (h *ConsistentHash) Get(v any) (any, bool) {
	h.lock.RLock()
	defer h.lock.RUnlock()

	if len(h.ring) == 0 {
		return nil, false
	}

	hash := h.hashFunc([]byte(repr(v))) // 根据查找键生成查找哈希键
	// 从哈希环中取出匹配的虚拟节点哈希键索引
	index := sort.Search(len(h.keys), func(i int) bool {
		return h.keys[i] >= hash
	}) % len(h.keys)

	nodes := h.ring[h.keys[index]]
	switch len(nodes) {
	case 0:
		return nil, false
	case 1:
		return nodes[0], true
	default:
		// 当多个真实节点映射到同一个虚拟节点时，使用二次哈希
		innerIndex := h.hashFunc([]byte(innerRepr(v)))
		pos := int(innerIndex % uint64(len(nodes)))
		return nodes[pos], true
	}
}

// Remove 删除节点 真实节点及其所有虚拟节点
func (h *ConsistentHash) Remove(node any) {
	nodeRepr := repr(node) // 将节点转换成字符串

	h.lock.Lock()
	defer h.lock.Unlock()

	if !h.containsNode(nodeRepr) {
		return
	}

	// 删除虚拟节点
	for i := 0; i < h.replicas; i++ {
		hash := h.hashFunc([]byte(nodeRepr + strconv.Itoa(i))) // 使用节点字符串+序号生成虚拟节点哈希键
		// 从哈希环中取出匹配的虚拟节点哈希键索引
		index := sort.Search(len(h.keys), func(i int) bool {
			return h.keys[i] >= hash
		})
		// 从keys中删除
		if index < len(h.keys) && h.keys[index] == hash {
			h.keys = append(h.keys[:index], h.keys[index+1:]...)
		}
		h.removeRingNode(hash, nodeRepr) // 从ring中删除
	}

	h.removeNode(nodeRepr) // 删除真实节点 从nodes中删除
}

// 从虚拟节点哈希键到真实节点的映射中删除真实节点
func (h *ConsistentHash) removeRingNode(hash uint64, nodeRepr string) {
	if nodes, ok := h.ring[hash]; ok {
		// 处理哈希冲突，即不同真实节点映射了相同的虚拟节点哈希键
		newNodes := nodes[:0]
		for _, x := range nodes {
			if repr(x) != nodeRepr {
				newNodes = append(newNodes, x)
			}
		}
		if len(newNodes) > 0 {
			// 有哈希冲突，将其它真实节点保留
			h.ring[hash] = newNodes
		} else {
			// 没有哈希冲突，删除虚拟节点
			delete(h.ring, hash)
		}
	}
}

func (h *ConsistentHash) addNode(nodeRepr string) {
	h.nodes[nodeRepr] = placeholder
}

func (h *ConsistentHash) containsNode(nodeRepr string) bool {
	_, ok := h.nodes[nodeRepr]
	return ok
}

func (h *ConsistentHash) removeNode(nodeRepr string) {
	delete(h.nodes, nodeRepr)
}

func innerRepr(node any) string {
	return fmt.Sprintf("%d:%v", prime, node)
}

func repr(node any) string {
	return Repr(node)
}

func Repr(v any) string {
	if v == nil {
		return ""
	}

	// if func (v *Type) String() string, we can't use Elem()
	switch vt := v.(type) {
	case fmt.Stringer:
		return vt.String()
	}

	val := reflect.ValueOf(v)
	for val.Kind() == reflect.Ptr && !val.IsNil() {
		val = val.Elem()
	}

	return reprOfValue(val)
}

func reprOfValue(val reflect.Value) string {
	switch vt := val.Interface().(type) {
	case bool:
		return strconv.FormatBool(vt)
	case error:
		return vt.Error()
	case float32:
		return strconv.FormatFloat(float64(vt), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(vt, 'f', -1, 64)
	case fmt.Stringer:
		return vt.String()
	case int:
		return strconv.Itoa(vt)
	case int8:
		return strconv.Itoa(int(vt))
	case int16:
		return strconv.Itoa(int(vt))
	case int32:
		return strconv.Itoa(int(vt))
	case int64:
		return strconv.FormatInt(vt, 10)
	case string:
		return vt
	case uint:
		return strconv.FormatUint(uint64(vt), 10)
	case uint8:
		return strconv.FormatUint(uint64(vt), 10)
	case uint16:
		return strconv.FormatUint(uint64(vt), 10)
	case uint32:
		return strconv.FormatUint(uint64(vt), 10)
	case uint64:
		return strconv.FormatUint(vt, 10)
	case []byte:
		return string(vt)
	default:
		return fmt.Sprint(val.Interface())
	}
}
