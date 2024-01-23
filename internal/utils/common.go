package utils

import (
	"context"
	"github.com/gogf/gf/v2/os/gfile"
	"log"
	"strings"
)

// 截取双引号里的文本
func SubStr(str string) (data string) {
	// 寻找第一个双引号的索引
	startIndex := strings.Index(str, "\"")
	if startIndex == -1 {
		log.Println("未找到双引号")
		return
	}

	// 寻找第二个双引号的索引
	endIndex := strings.Index(str[startIndex+1:], "\"")
	if endIndex == -1 {
		log.Println("未找到第二个双引号")
		return
	}

	// 截取双引号里面的内容
	content := str[startIndex+1 : startIndex+1+endIndex]

	return content
}

// 创建文件
func CreateFile(path string) error {

	exists := gfile.Exists(path)
	if !exists {
		_, err := gfile.Create(path)
		if err != nil {
			return err
		}
	}
	return nil
}

// post请求
func HttpPost(ctx context.Context) error {

	return nil
}

// get请求
func HttpGet(ctx context.Context) error {

	return nil
}
