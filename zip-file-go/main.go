package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/sqweek/dialog"
)

func main() {
	needDirectory := dialog.Message("需要一个文件夹？").Title("File Or Directory").YesNo()

	var fileString string
	var err error
	if needDirectory {
		fileString, err = dialog.Directory().Title("选择要复制的文件夹").Browse()
	} else {
		fileString, err = dialog.File().Title("选择要复制的文件").Load()
	}
	if err != nil {
		fmt.Println("An error occured", err)
		return
	}
	dialog.Message("你选择了\"" + fileString + "\"，现在选择要复制到的文件夹").Info()
	var copyToString string
	copyToString, err = dialog.Directory().Title("选择要复制到的文件夹").Browse()
	if err != nil {
		fmt.Println("An error occured", err)
		return
	}
	dialog.Message("即将复制到" + copyToString).Info()

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		panic(err)
	}
	fmt.Println(l.Addr().String())

	http.HandleFunc("/download", downloadFileHandler)
	http.HandleFunc("/shutdown", shutdownHandler)
	
	args := os.Args
	exePath := path.Join(filepath.Dir(args[0]), "go-download-client.exe")
	clientArgs := []string{l.Addr().String(), fileString, copyToString}
	cmd := exec.Command(exePath, clientArgs...)

	// 启动服务器
    srv := &http.Server{
        Addr: l.Addr().String(),
    }

	go func() {
		err = srv.Serve(l)
		if err != nil {
			panic(err)
		}	
	}()
		time.Sleep(2 * time.Second)
		err = cmd.Start()
		if err != nil {
			panic(err)
		}

	// 等待信号以优雅关闭服务器
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    // 关闭服务器
    fmt.Println("Shutting down server...")
    if err := srv.Shutdown(nil); err != nil {
        fmt.Println("Server Shutdown:", err)
    }
    fmt.Println("Server exited")
}

func downloadFileHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
		requestFilePath := query.Get("request_file_path")
		fmt.Println(requestFilePath)

		fileName := path.Base(requestFilePath)
		w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
		
		http.ServeFile(w, r, requestFilePath)
}

func shutdownHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Println("Shutdown requested")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Server is shutting down..."))

    // 关闭服务器
    go func() {
        dialog.Message("操作已经全部完成").Info()
        // 退出程序
        os.Exit(0)
    }()
}
