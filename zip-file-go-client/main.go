package main

import (
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
)

func main() {
	args := os.Args

	if len(args) < 4 {
		panic("args is not enough")
	}

	serverUrl := args[1]
	selectedFile := args[2]
	savePath := args[3]

	// todo 支持文件夹遍历下载
	if isPathDir(selectedFile) {
		// 遍历保存
		err := filepath.Walk(selectedFile, func(walkingPath string, info fs.FileInfo, err error) error {
			if err != nil {
				return err
			}
			fmt.Println("Walking ", walkingPath)
			if isPathDir(walkingPath) {
				return nil
			}
			relativePath, err := filepath.Rel(selectedFile, walkingPath)
			if err != nil {
				panic(err)
			}
			err = downloadFile(serverUrl, walkingPath, filepath.Join(savePath, relativePath))
			if err != nil {
				fmt.Println("Download Error", err)
				return nil
			}
			return nil
		})
		if err != nil {
			panic(err)
		}
	} else {
		fileName := filepath.Base(selectedFile)
		err := downloadFile(serverUrl, selectedFile, filepath.Join(savePath, fileName))
		if err != nil {
			fmt.Println("Download Error", err)
		}
	}
	shutdownUrl := fmt.Sprintf("http://%s/shutdown", serverUrl)
	_, err := http.Get(shutdownUrl)
	if err != nil {
		panic(err)
	}
}

func isPathDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return false
	}
	return info.IsDir()
}

func downloadFile(serverAddr, filePath string, savePath string) error {
	// 构建请求的 URL
	url := fmt.Sprintf("http://%s/download?request_file_path=%s", serverAddr, filePath)

	// 发送 GET 请求
	response, err := http.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close() // 确保在函数结束时关闭响应体

	// 检查响应状态
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", response.StatusCode)
	}

	// 创建文件
	dir := filepath.Dir(savePath)
	// 确保目录存在，如果不存在则创建它
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Println("创建目录时出错:", err)
		return err
	}
	fmt.Println(savePath)
	outFile, err := os.Create(savePath + ".data") // 指定保存的文件名
	if err != nil {
		return err
	}
	defer outFile.Close() // 确保在函数结束时关闭文件

	// 将响应体写入文件
	_, err = io.Copy(outFile, response.Body)
	return err
}
