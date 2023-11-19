package oldstuff

/*
import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func main() {
	mapa := getEnvironmentVariables2()

	for key, value := range mapa {
		fmt.Printf("Key: %d, Value: %d \n", key, value)
	}

}

func getEnvironmentVariables2() (updatedEnv map[int]int) {
	env := os.Environ()
	envSorted := []string{}
	envMap := make(map[int]int)

	for i := 0; i < len(env); i++ {
		if strings.Contains(env[i], "Node_") {
			envSorted = append(envSorted, env[i])
		}
	}

	for i := 0; i < len(envSorted); i++ {
		current := strings.TrimPrefix(envSorted[i], "Node_")
		array := strings.Split(current, "=")

		id, _ := strconv.Atoi(array[0])
		port, _ := strconv.Atoi(array[1])

		envMap[id] = port
	}

	return envMap
}
*/
