package structure

// 参考网址：https://blog.csdn.net/qq_26981997/article/details/78773487

import (
	"sync/atomic"
	"unsafe"
)

// 单向链表
type Node struct {
	next *Node
	data interface{}
}

type Queue struct {
	head , tail *Node
	isUP atomic.Value
	single chan struct{}
}

func New() *Queue {
	q := &Queue{}
	q.head = &Node{}
	q.tail = q.head

	q.isUP = atomic.Value{}
	q.isUP.Store(false)
	q.single = make(chan struct{}, 1)

	return q
}

func (qu *Queue)Single()<-chan struct{}{
	return qu.single
}

func (qu *Queue)Push(x interface{}){
	n := &Node{data: x}
	prev := (*Node)(atomic.SwapPointer((*unsafe.Pointer)(unsafe.Pointer(&qu.head)), unsafe.Pointer(n)))
	prev.next = n
	// qu.SingleUP()
}

func (qu *Queue)SingleUP(force bool){
	if force{
		qu.single <- struct{}{}
	} else if ! qu.isUP.Load().(bool){
		qu.isUP.Store(true)
		qu.single <- struct{}{}
	}
}

// 需要注意的是，这个函数要在处理线程处理完后再去调用
// 否则会造成写数据阻塞
func (qu *Queue) SingleDown() {
	qu.isUP.Store(false)
}

func (qu *Queue)Pop()interface{}  {
	tail := qu.tail
	next := (*Node)(atomic.LoadPointer((*unsafe.Pointer)(unsafe.Pointer(&tail.next)))) // acquire
	if next != nil {
		qu.tail = next
		v := next.data
		next.data = nil
		return v
	}//else {
	//	qu.isUP.Store(false)
	//}

	return nil
}

func (qu *Queue)Close()  {
	close(qu.single)
}