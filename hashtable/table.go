/**
 * Copyright 2017 tsf Author. All Rights Reserved.
 * email: donnie4w@gmail.com
 */
package hashtable

import (
	"sync"
)

type HashTable struct {
	maptable map[interface{}]interface{}
	lock     *sync.RWMutex
}

func NewHashTable() *HashTable {
	return &HashTable{make(map[interface{}]interface{}, 0), new(sync.RWMutex)}
}

func (this *HashTable) Put(k, v interface{}) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.maptable[k] = v
}

func (this *HashTable) Get(k interface{}) (i interface{}) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	i = this.maptable[k]
	return
}

func (this *HashTable) Exist(k interface{}) bool {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if _, ok := this.maptable[k]; !ok {
		return true
	}
	return false
}

/**all keys */
func (this *HashTable) GetKeys() (keys []interface{}) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	keys = make([]interface{}, 0)
	for k, _ := range this.maptable {
		keys = append(keys, k)
	}
	return
}

/**all values*/
func (this *HashTable) GetValues() (values []interface{}) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	values = make([]interface{}, 0)
	for _, v := range this.maptable {
		values = append(values, v)
	}
	return
}

/** put if key not exist */
func (this *HashTable) Putnx(k, v interface{}) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if _, ok := this.maptable[k]; !ok {
		this.maptable[k] = v
	}
}

func (this *HashTable) Del(k interface{}) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.maptable, k)
}

/**delete if key and value exist*/
func (this *HashTable) Delnx(k, v interface{}) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	if value, ok := this.maptable[k]; ok {
		if value == v {
			delete(this.maptable, k)
			return true
		}
	}
	return false
}
