package bcg

import (
	"os"
	"path/filepath"
)

func ReadFile(fn string) []byte {
	file, err := os.Open(fn)
	if err != nil {
		//LogRed(err.Error())
		return nil
	}
	defer file.Close()
	var data []byte
	fi, err := os.Stat(fn)
	data = make([]byte, fi.Size())
	_, err = file.Read(data)
	if err != nil {
		LogRed(err.Error())
		return nil
	}
	return data
}
func FileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// CreateFolder 生成路径指定的文件夹，包括所有上级文件夹
func CreateFolder(folder string) error {
	err := os.MkdirAll(folder, 0777)
	if err != nil {
		LogRed(err.Error())
	}
	return err
}

// SaveFileAndFolder 函数会自动创建目标文件所在的目录文件夹，如果目标文件已经存在，数据会被覆盖，
// 此函数和 SaveFileToFolder 的区别是无需分开传输目录和文件参数
func SaveFileAndFolder(fn string, data []byte) (err error) {
	folder := filepath.Dir(fn)
	err = os.MkdirAll(folder, 0777)
	if CheckError(err) {
		return
	}
	return SaveFile(fn, data)
}

// SaveFileToFolder 函数会自动创建目标文件所在的目录文件夹，如果目标文件已经存在，数据会被覆盖
func SaveFileToFolder(folder, fn string, data []byte) (err error) {
	err = os.MkdirAll(folder, 0777)
	if err != nil {
		LogRed(err.Error())
		return err
	}
	folder += "/" + fn
	return SaveFile(folder, data)
}

// SaveFileToFolderExistFail 保存到文件，并且生成所有上级文件夹（如果不存在），但是如果文件存在，函数失败，
// 用于防止已经存在的文件被覆盖
func SaveFileToFolderExistFail(folder, fn string, data []byte) (err error) {
	err = os.MkdirAll(folder, 0777)
	if err != nil {
		LogRed(err.Error())
		return err
	}
	folder += "/" + fn
	return SaveFileExistFail(folder, data)
}

//SaveFileExistFail 防止同名文件被覆盖，只有目标文件不存在才会成功
func SaveFileExistFail(fn string, data []byte) (err error) {
	file, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0777)
	if err != nil {
		LogRed(err.Error())
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		LogRed(err.Error())
		return err
	}
	return err
}

// SaveFile 函数保存数据到指定文件，要求目标目录必须存在，否则失败。目标文件如果已经存在，会被新数据覆盖
func SaveFile(fn string, data []byte) (err error) {
	file, err := os.OpenFile(fn, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		LogRed(err.Error())
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		LogRed(err.Error())
		return err
	}
	return err
}
