package lruMap

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type lruNode struct {
	key        string
	value      interface{}
	prov, next *lruNode
	timestamp  int64
}

type Manager struct {
	ctx        context.Context
	cancel     func()
	max, count int
	root, last *lruNode
	value      *sync.Map
	setCh      chan *lruNode
}

func New(size int) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	m := Manager{
		ctx:    ctx,
		cancel: cancel,
		max:    size,
		count:  0,
		root:   nil,
		last:   nil,
		value:  new(sync.Map),
		setCh:  make(chan *lruNode, 200),
	}

	go m.dispose()
	return &m
}

func (m *Manager) Close() {
	close(m.setCh)
	m.cancel()
	m.value = nil
}

func (m *Manager) Get(key string) (interface{}, bool) {
	v, ok := m.value.Load(key)
	if !ok {
		return nil, false
	}

	info, ok := v.(*lruNode)
	if !ok {
		return nil, false
	}

	return info.value, true
}

func (m *Manager) Set(key string, value interface{}) {
	if nil == m.value {
		return
	}

	m.setCh <- &lruNode{
		key:   key,
		value: value,
	}
}

func (m *Manager) dispose() {
	var (
		ok        bool
		loadValue interface{}
		recv, old *lruNode
	)

	for {
		select {
		case recv = <-m.setCh:
			{

				loadValue, ok = m.value.Load(recv.key)

				// set to update times
				if ok {
					old = loadValue.(*lruNode)
					old.value = recv.value
					update(m, old)
					continue
				}

				if m.count >= m.max {
					m.value.Delete(m.last.key)
					m.deleteNode(m.last)
				} else {
					m.count++
				}

				recv.timestamp = time.Now().Unix()

				m.value.Store(recv.key, recv)
				m.insertNode(recv)
			}
		case <-m.ctx.Done():
			{
				return
			}
		}
	}
}

func update(m *Manager, node *lruNode) {
	node.timestamp = time.Now().Unix()
	m.deleteNode(node)
	m.insertNode(node)
}

func (m *Manager) insertNode(now *lruNode) {
	// init
	if nil == m.root {
		m.root = now
		m.last = now
		return
	}

	var temp = m.root
	for nil != temp {
		if temp.timestamp <= now.timestamp {
			break
		}

		temp = temp.next
	}

	// replace last
	if nil == temp {
		m.last.next = now
		now.prov = m.last
		now.next = nil
		m.last = now
		return
	}

	// replace root
	if nil == temp.prov {
		now.prov = nil
		now.next = temp
		temp.prov = now
		m.root = now
		return
	}

	// insert temp before
	var (
		prov = temp.prov
	)

	prov.next = now
	temp.prov = now

	now.prov = prov
	now.next = temp
}

func (m *Manager) deleteNode(now *lruNode) {
	defer func() {
		now.prov = nil
		now.next = nil
	}()

	// only one node
	if nil == now.prov && nil == now.next {
		m.root = nil
		m.last = nil
		return
	}

	// is last
	if nil == now.next {
		now.prov.next = nil
		m.last = now.prov
		return
	}

	// is root
	if nil == now.prov {
		now.next.prov = nil
		m.root = now.next
		return
	}

	var (
		prov = now.prov
		next = now.next
	)

	prov.next = next
	next.prov = prov
}

func RangePrint(m *Manager) {
	var temp = m.root
	for nil != temp {
		fmt.Println(temp.key, temp.value, temp.timestamp)
		temp = temp.next
	}

	fmt.Println()
}
