package shared

import "sync"

var (
	globalId    int
	globalMutex sync.Mutex
)

func SetGlobalId(id int) {
	globalMutex.Lock()
	globalId = id
	globalMutex.Unlock()
}

func GetGlobalId() int {
	globalMutex.Lock()
	defer globalMutex.Unlock()
	return globalId
}
