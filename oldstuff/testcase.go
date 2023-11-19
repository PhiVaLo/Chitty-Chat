package oldstuff

/*
func main() {
	os.Setenv("Node_1", "2555")
	os.Setenv("Node_2", "2556")
	os.Setenv("Node_3", "2557")
	os.Setenv("Node_4", "2558")

	mapa := getEnvironmentVariables()

	for key, value := range mapa {
		fmt.Printf("Key: %d, Value: %d \n", key, value)
	}

	for {

	}
}

func getEnvironmentVariables() (updatedEnv map[int]int) {
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
