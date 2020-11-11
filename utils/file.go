package utils

import (
	"io/ioutil"
	"os"
)

func CheckFileExists(path string) bool {
	stat, err := os.Stat(path)
	return !os.IsNotExist(err) && !stat.IsDir()
}

func WriteFile(filename string, data []byte, perm os.FileMode) error {
	data = append(data, '\n')
	return ioutil.WriteFile(filename, data, perm)
}
