package file

import (
	"fmt"
	"os"
)

func GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("file does not exist: %s", filePath)
		}
		return 0, fmt.Errorf("unable to access file %s: %w", filePath, err)
	}

	return info.Size(), nil
}