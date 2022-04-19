package utils

import (
	"io/fs"
	"os"
)

// 创建目录
func CreateDir(dir string, perm fs.FileMode) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			err = os.MkdirAll(dir, perm)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}
	return nil
}
