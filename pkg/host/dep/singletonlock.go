package dep

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type SingletonLock interface {
	// return false - re-entrance lock, true - locked
	Lock() bool
	// return false - re-entrance unlock, true - unlocked
	Unlock() bool
}

type ThreadSafeLock struct {
	compName  string
	gid       int32
	reentrant int32
	mutex     *sync.Mutex
}

// thread safe lock with re-entrant support
func NewThreadSafeLock(compName string) *ThreadSafeLock {
	mutex := &sync.Mutex{}
	return &ThreadSafeLock{
		compName:  compName,
		reentrant: 0,
		gid:       0,
		mutex:     mutex,
	}
}
func (tsl *ThreadSafeLock) IsLocked() bool {
	return atomic.LoadInt32(&tsl.reentrant) > 0
}

func (tsl *ThreadSafeLock) lockWithGId(current_gid int32) {
	tsl.mutex.Lock()
	//fmt.Printf("locked by %d\n", current_gid)
	atomic.CompareAndSwapInt32(&tsl.gid, 0, current_gid)
}
func (tsl *ThreadSafeLock) Lock() bool {
	current_gid := goroutine_id()
	locked_gid := atomic.LoadInt32(&tsl.gid)
	//fmt.Printf("goroutine %d try locking on singleton %s, current locked by %d\n", current_gid, tsl.compName, locked_gid)
	if locked_gid == current_gid {
		atomic.AddInt32(&tsl.reentrant, 1)
		return false
	} else {
		tsl.lockWithGId(current_gid)
		atomic.StoreInt32(&tsl.reentrant, 1)
		return true
	}
}
func (tsl *ThreadSafeLock) Unlock() bool {
	current_gid := goroutine_id()
	locked_gid := atomic.LoadInt32(&tsl.gid)
	if locked_gid == current_gid {
		reentrant := atomic.AddInt32(&tsl.reentrant, -1)
		if reentrant <= 0 {
			atomic.CompareAndSwapInt32(&tsl.gid, current_gid, 0)
			tsl.mutex.Unlock()
			return true
		}
		return false
	} else {
		panic(fmt.Errorf("cannot unlock a lock owned by other goroutine %d: component - %s, current goroutine - %d", locked_gid, tsl.compName, current_gid))
	}
}

func goroutine_id() int32 {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil || id == 0 {
		panic(fmt.Sprintf("cannot get goroutine id: %v, %d", err, id))
	}
	return int32(id)
}
