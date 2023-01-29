package et

import "sync"

type RWLocker interface {
	sync.Locker
	RLock()
	RUnlock()
}
