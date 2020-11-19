package lb

import (
	"encoding/json"
	"math/rand"
	"sync"
)

/*
	并发安全的map，采用分段加读写锁的方法
	copy from https://github.com/orcaman/concurrent-map
*/
const (
	min_shard = 5
	max_shard = 64
)
var (
	hash_table = []uint32{998244353, 131 , 16777619, 19190817}
)
var(
	//分段大小
	SHARD_COUNT = 32
)
type ConcurrentMap struct{
	Shareds []*ConcurrentMapShared
}

type ConcurrentMapShared struct{
	items   map[string]interface{}
	sync.RWMutex
}
//最好设置成32
func New(shard int) *ConcurrentMap{
	if shard > max_shard || shard < min_shard{
		shard = SHARD_COUNT
	}
	SHARD_COUNT = shard
	m := make([]*ConcurrentMapShared, SHARD_COUNT)
	for i := 0 ; i < SHARD_COUNT ; i ++ {
		m[i] = &ConcurrentMapShared{items: make(map[string]interface{})}
	}
	return &ConcurrentMap{Shareds: m}
}
func fnv32(key string) uint32 {
	hash := uint32(2166136261)
	var prime32 uint32
	for i := 0; i < len(key); i++ {
		prime32 = hash_table[uint32(rand.Intn(4))]
		hash *= prime32
		hash ^= uint32(key[i])
	}
	return hash
}
// GetShard returns shard under given key
func (m *ConcurrentMap) GetShard(key string) *ConcurrentMapShared {
	return m.Shareds[uint(fnv32(key))%uint(SHARD_COUNT)]
}

func (m *ConcurrentMap) MapSet(data map[string]interface{}) {
	for key, value := range data {
		shard := m.GetShard(key)
		shard.Lock()
		shard.items[key] = value
		shard.Unlock()
	}
}

// Sets the given value under the specified key.
func (m *ConcurrentMap) Set(key string, value interface{}) {
	// Get map shard.
	shard := m.GetShard(key)
	shard.Lock()
	shard.items[key] = value
	shard.Unlock()
}

// Get retrieves an element from map under given key.
func (m *ConcurrentMap) Get(key string) (interface{}, bool) {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	// Get item from shard.
	val, ok := shard.items[key]
	shard.RUnlock()
	return val, ok
}

// Count returns the number of elements within the map.
func (m *ConcurrentMap) Count() int {
	count := 0
	for i := 0; i < SHARD_COUNT; i++ {
		shard := m.Shareds[i]
		shard.RLock()
		count += len(shard.items)
		shard.RUnlock()
	}
	return count
}

// Looks up an item under specified key
func (m *ConcurrentMap) Has(key string) bool {
	// Get shard
	shard := m.GetShard(key)
	shard.RLock()
	// See if element is within shard.
	_, ok := shard.items[key]
	shard.RUnlock()
	return ok
}

// Remove removes an element from the map.
func (m *ConcurrentMap) Remove(key string) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	delete(shard.items, key)
	shard.Unlock()
}

// Pop removes an element from the map and returns it
func (m *ConcurrentMap) Pop(key string) (v interface{}, exists bool) {
	// Try to get shard.
	shard := m.GetShard(key)
	shard.Lock()
	v, exists = shard.items[key]
	delete(shard.items, key)
	shard.Unlock()
	return v, exists
}
// IsEmpty checks if map is empty.
func (m ConcurrentMap) IsEmpty() bool {
	return m.Count() == 0
}

//采用协程与channel遍历map
//遍历返回的结果
type Tuple struct{
	Key string
	Val interface{}
}

//用于for 遍历map 的iterator , 采用带缓存的channel 实现
func (m *ConcurrentMap) IterBuffered () <- chan Tuple{
	chs , total := m.snapshot()
	ch := make(chan Tuple , total)
	go m.readChannel(chs , ch)  //这里用协程去读，因为在最外面的items返回所有的数据时，也会从channel中阻塞的读取数据
	return ch
}

//制作ConcurrentMap的快照，相当于深拷贝
func (m *ConcurrentMap) snapshot () ([]chan Tuple, int){
	chans := make([]chan Tuple , SHARD_COUNT)
	var wg sync.WaitGroup
	total := 0
	wg.Add(SHARD_COUNT)

	for index := range m.Shareds{
		go func(i int){
			m.Shareds[i].RLock()
			chans[i] = make(chan Tuple, len(m.Shareds[i].items))
			total += len(m.Shareds[i].items)
			defer wg.Done()
			for key , val := range m.Shareds[i].items{
				chans[i] <- Tuple{
					Key: key,
					Val: val,
				}
			}
			m.Shareds[i].RUnlock()
			close(chans[i])
		}(index)
	}
	wg.Wait()
	return chans, total
}
//从snapshot返回的channels中读取所有的元素，放入 'out'中
func (m *ConcurrentMap) readChannel(chans []chan Tuple , out chan Tuple){
	var wg sync.WaitGroup
	wg.Add(len(chans))
	for _ , c := range chans{
		go func(){
			for v := range c{
				out <- v
			}
			defer wg.Done()
		}()
	}
	wg.Wait()
	close(out)
}

// Items 将以map[string]interface{}的形式返回所有的元素
func (m *ConcurrentMap) Items() map[string]interface{} {
	tmp := make(map[string]interface{})

	// Insert items to temporary map.
	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}
	return tmp
}


// Keys 将以[]string的形式返回所有的keys
func (m *ConcurrentMap) Keys() []string {
	count := m.Count()
	ch := make(chan string, count)
	//以协程的方式读key
	go func() {
		// Foreach shard.
		wg := sync.WaitGroup{}
		wg.Add(SHARD_COUNT)
		for _, shard := range m.Shareds {
			go func(shard *ConcurrentMapShared) {
				defer wg.Done()
				shard.RLock()
				for key := range shard.items {
					ch <- key
				}
				shard.RUnlock()
			}(shard)
		}
		wg.Wait()
		close(ch)
	}()

	// Generate keys
	keys := make([]string, 0, count)
	for k := range ch {
		keys = append(keys, k)
	}
	return keys
}
//Reviles ConcurrentMap "private" variables to json marshal.
func (m *ConcurrentMap) MarshalJSON() ([]byte, error) {
	// Create a temporary map, which will hold all item spread across shards.
	tmp := make(map[string]interface{})

	// Insert items to temporary map.
	for item := range m.IterBuffered() {
		tmp[item.Key] = item.Val
	}
	return json.Marshal(tmp)
}

