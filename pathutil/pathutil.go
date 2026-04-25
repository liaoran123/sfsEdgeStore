package pathutil

import (
	"os"
	"path/filepath"
	"sync"
)

var (
	exeDir string
	once   sync.Once
	pathErr error
)

// GetExecutableDir 获取可执行文件所在目录
func GetExecutableDir() (string, error) {
	once.Do(func() {
		path, e := os.Executable()
		if e != nil {
			pathErr = e
			return
		}
		exeDir = filepath.Dir(path)
	})
	return exeDir, pathErr
}

// Join 构建相对于可执行文件目录的路径
func Join(parts ...string) (string, error) {
	dir, e := GetExecutableDir()
	if e != nil {
		return "", e
	}
	return filepath.Join(append([]string{dir}, parts...)...), nil
}
