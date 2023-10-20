// package shared

// import "sync"

// var (
// 	globalID    int
// 	globalMutex sync.Mutex
// )

// func init() {
//     // Initialize the global ID when the shared package is initialized
//     globalID = 1 // or any initial value you prefer
// }

// func SetGlobalID(id int) {
//     globalMutex.Lock()
//     globalID = id
//     globalMutex.Unlock()
// }

// func GetGlobalID() int {
//     globalMutex.Lock()
//     defer globalMutex.Unlock()
//     return globalID
// }

// func IncrementGlobalID() {
//     globalMutex.Lock()
//     globalID++
//     globalMutex.Unlock()
// }


package shared

import "sync"

var (
    globalID    int
    globalMutex sync.Mutex
)

func GenerateUniqueClientID() int {
    globalMutex.Lock()
    globalID++
    newID := globalID
    globalMutex.Unlock()
    return newID
}