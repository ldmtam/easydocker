package easydocker

import (
	"os"
)

func getCurrentDir() string {
	currentDir, _ := os.Getwd()
	return currentDir
}
