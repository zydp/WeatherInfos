package lrucache

import "container/list"

/* Cache is an LRU cache. It is not safe for concurrent access.*/
type Cache struct {
	MaxEntries int                              /*最大数量限制，0表示不限制*/
	OnEvicted  func(key Key, value interface{}) /*驱逐回调，从缓存清除条目时执行*/
	ll         *list.List                       /*用于lru管理的list*/
	cache      map[interface{}]*list.Element    /*缓存*/
}

/*key是可以比较的任意类型*/
type Key interface{}

/*缓存的k/v对*/
type entry struct {
	key   Key
	value interface{}
}

/*创建cache*/
func New(maxEntries int) *Cache {
	return &Cache{
		MaxEntries: maxEntries,
		ll:         list.New(),
		cache:      make(map[interface{}]*list.Element),
	}
}

/*向cache添加一个键值对*/
func (c *Cache) Add(key Key, value interface{}) {
	if c.cache == nil { //安全检查
		c.cache = make(map[interface{}]*list.Element)
		c.ll = list.New()
	}
	if ee, ok := c.cache[key]; ok { //如果存在，则在链表中将ee插入到最前，并更新ee的值  set a 123; set a 456
		c.ll.MoveToFront(ee) //链表的基本操作，1、使用删除获取该元素的地址 2、重新将获取的地址插入到链表最前
		ee.Value.(*entry).value = value
		return
	}
	//如果不存在，则创建新的内部对象，并放到最前
	//链表中保存内部k/v对象的地址，并返回在链表中的元素地址
	ele := c.ll.PushFront(&entry{key, value})

	//缓存中保存了 [key]内部k/v对象在链表中的元素地址
	c.cache[key] = ele

	//判断缓存数量，是否超过最大值，如果超过，则移出旧的对象（链表末尾表示所有缓存对象中，最长时间未被使用）
	if c.MaxEntries != 0 && c.ll.Len() > c.MaxEntries {
		c.RemoveOldest()
	}
}

/*获取时，如果元素存在，则将元素的位置在链表中更新*/
func (c *Cache) Get(key Key) (value interface{}, ok bool) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return
}

/*从缓存中删除一个元素*/
func (c *Cache) Remove(key Key) {
	if c.cache == nil {
		return
	}
	if ele, hit := c.cache[key]; hit {
		c.removeElement(ele)
	}
}

/*移除旧的对象（链表末尾表示所有缓存对象中，最长时间未被使用）*/
func (c *Cache) RemoveOldest() {
	if c.cache == nil {
		return
	}
	ele := c.ll.Back() //获取链表尾
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *Cache) removeElement(e *list.Element) {
	c.ll.Remove(e) //从链表中移除旧对象
	kv := e.Value.(*entry)
	delete(c.cache, kv.key) //从缓存中删除记录
	if c.OnEvicted != nil {
		c.OnEvicted(kv.key, kv.value) //删除时回调
	}
}

/*获取长度*/
func (c *Cache) Len() int {
	if c.cache == nil {
		return 0
	}
	return c.ll.Len()
}

/*清空整个缓存*/
func (c *Cache) Clear() {
	if c.OnEvicted != nil {
		for _, e := range c.cache {
			kv := e.Value.(*entry)
			c.OnEvicted(kv.key, kv.value)
		}
	}
	c.ll = nil
	c.cache = nil
}
