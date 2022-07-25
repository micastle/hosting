package dep

import (
	"math/rand"
	"testing"
	"time"

	"goms.io/azureml/mir/mir-vmagent/pkg/host/test"
)

func TestLock_basic(t *testing.T) {
	lock := NewThreadSafeLock("MyComp")
	if lock.IsLocked() {
		t.Error("initial lock should not be locked already")
	}

	if !lock.Lock() {
		t.Error("first lock should success and return non re-entrance")
	}

	if !lock.IsLocked() {
		t.Error("lock status should be true after the first lock")
	}

	if !lock.Unlock() {
		t.Error("first unlock after first lock should success and return non re-entrance")
	}

	if lock.IsLocked() {
		t.Error("lock status should be false after unlock")
	}
}

func TestLock_re_entrance(t *testing.T) {
	lock := NewThreadSafeLock("MyComp")

	lock.Lock()
	if lock.Lock() {
		t.Error("second lock should success and return re-entrance")
	}

	if !lock.IsLocked() {
		t.Error("lock status should be true after re-entrance lock")
	}

	lock.Unlock()
	if !lock.IsLocked() {
		t.Error("lock status should be true after one unlock on re-entrance lock")
	}

	if !lock.Unlock() {
		t.Error("first unlock after first lock should success and return non re-entrance")
	}

	if lock.IsLocked() {
		t.Error("lock status should be false after totally unlocked from re-entrance lock")
	}
}

func TestLock_unlock_not_locked(t *testing.T) {
	defer test.AssertPanicContent(t, "cannot unlock a lock owned by other goroutine 0:", "panic content not expected")

	lock := NewThreadSafeLock("MyComp")
	if lock.IsLocked() {
		t.Error("initial lock should not be locked already")
	}

	lock.Unlock() // panic
}

func TestLock_unlock_more(t *testing.T) {
	defer test.AssertPanicContent(t, "cannot unlock a lock owned by other goroutine 0:", "panic content not expected")

	lock := NewThreadSafeLock("MyComp")
	if lock.IsLocked() {
		t.Error("initial lock should not be locked already")
	}

	lock.Lock()
	lock.Unlock()
	if lock.IsLocked() {
		t.Error("lock status should be false after unlocked")
	}

	lock.Unlock() // panic
}

func TestLock_unlock_not_owner(t *testing.T) {
	defer test.AssertPanicContent(t, "cannot unlock a lock owned by other goroutine", "panic content not expected")

	lock := NewThreadSafeLock("MyComp")
	if lock.IsLocked() {
		t.Error("initial lock should not be locked already")
	}

	done := make(chan bool, 1)
	go func() {
		lock.Lock()
		done <- true
	}()

	<-done

	lock.Unlock() // panic
}

func TestLock_concurrent(t *testing.T) {
	lock := NewThreadSafeLock("MyComp")

	start := make(chan bool)
	stop := false
	size := 10
	complete := make(chan bool, size)

	rand.Seed(time.Now().UnixNano())

	startLockOperator := func() {
		<-start

		for !stop {
			reentrant := rand.Intn(5) + 1
			for i := 0; i < reentrant; i++ {
				lock.Lock()
			}
			for i := 0; i < reentrant; i++ {
				lock.Unlock()
			}
		}

		complete <- true
	}

	for i := 0; i < size; i++ {
		go startLockOperator()
	}

	// align and start all goroutines
	time.Sleep(300 * time.Millisecond)
	close(start)
	// execute for a period and stop all goroutines
	time.Sleep(3 * time.Second)
	stop = true

	done := false
	go func() {
		for i := 0; i < size; i++ {
			<-complete
		}
		done = true
	}()

	time.Sleep(500 * time.Millisecond)
	if !done {
		t.Errorf("not done in time, dead locked?")
	}
}
