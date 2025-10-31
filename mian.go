package main

import (
	"fmt"
	"github.com/gogf/gf/v2/os/gcmd"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// 处理上传页面和文件上传请求
func uploadHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// 返回上传页面
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>文件上传</title>
    <style>
        body { font-family: Arial, sans-serif; padding: 20px; }
        input[type="file"] { margin-bottom: 10px; }
        button { padding: 8px 16px; font-size: 16px; }
    </style>
</head>
<body>
    <h2>上传文件</h2>
    <form action="/" method="post" enctype="multipart/form-data">
        <input type="file" name="uploadfile[]" multiple required />
        <br />
        <button type="submit">上传</button>
    </form>
</body>
</html>
        `)

	case http.MethodPost:
		// 解析表单
		err := r.ParseMultipartForm(10 << 30) // 限制最大上传 10GB
		if err != nil {
			http.Error(w, "解析表单失败: "+err.Error(), http.StatusBadRequest)
			return
		}

		files := r.MultipartForm.File["uploadfile[]"]
		if len(files) == 0 {
			http.Error(w, "没有上传文件", http.StatusBadRequest)
			return
		}

		// 创建保存目录
		err = os.MkdirAll("uploads", os.ModePerm)
		if err != nil {
			http.Error(w, "创建目录失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		var savedFiles []string
		for _, header := range files {
			file, err := header.Open()
			if err != nil {
				http.Error(w, "打开上传文件失败: "+err.Error(), http.StatusBadRequest)
				return
			}
			defer file.Close()

			filename := strconv.FormatInt(time.Now().UnixMilli(), 10) + "--" + filepath.Base(header.Filename)
			dstPath := filepath.Join("uploads", filename)

			dst, err := os.Create(dstPath)
			if err != nil {
				http.Error(w, "保存文件失败: "+err.Error(), http.StatusInternalServerError)
				return
			}
			defer dst.Close()

			_, err = io.Copy(dst, file)
			if err != nil {
				http.Error(w, "写入文件失败: "+err.Error(), http.StatusInternalServerError)
				return
			}

			savedFiles = append(savedFiles, header.Filename)

			log.SetFlags(0)
			log.Printf("%s，保存文件成功：%s\n", time.Now().Format(time.DateTime), filename)
		}

		// 返回上传成功页面
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w,
			`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>上传成功</title>
</head>
<body>
    <h2>文件上传成功！</h2>
    <ul>
`)
		for _, name := range savedFiles {
			fmt.Fprintf(w, "<li>%s</li>\n", name)
		}
		fmt.Fprint(w, `
    </ul>
    <a href="/">返回上传页面</a>
</body>
</html>`)

	default:
		http.Error(w, "不支持的请求方法", http.StatusMethodNotAllowed)
	}
}

func main() {
	http.HandleFunc("/", uploadHandler)

	// 命令，命令参数
	fmt.Println(gcmd.GetArgAll(), gcmd.GetOptAll())

	// 设置端口
	port := gcmd.GetOptWithEnv("port", 8080).String()
	port = gcmd.GetOptWithEnv("p", 8080).String()

	fmt.Println("服务器启动，监听端口 " + port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		fmt.Println("服务器启动失败:", err)
	}
}
