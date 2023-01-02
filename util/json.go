package util

import (
	"encoding/json"
	"fmt"

	"io/ioutil"
	"os"
)

// 结构体转为json
func Struct2Json(obj interface{}) string {
	str, err := json.Marshal(obj)
	if err != nil {
		panic(fmt.Sprintf("[Struct2Json]转换异常: %v", err))
	}
	return string(str)
}

// json转为结构体
func Json2Struct(str string, obj interface{}) {
	// 将json转为结构体
	err := json.Unmarshal([]byte(str), obj)
	if err != nil {
		panic(fmt.Sprintf("[Json2Struct]转换异常: %v", err))
	}
}

// json interface转为结构体
func JsonI2Struct(str interface{}, obj interface{}) {
	JsonStr := str.(string)
	Json2Struct(JsonStr, obj)
}

// 获取conf文件目录
func getConfFile(fileName string) string {
	workDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Errorf("read program runtime dir fail:%s \n", err))
	}
	return workDir + "/conf/" + fileName
}

// conf目录下的文件(自动添加.json后缀)转换为指定结构体
func ConfFileToModel(fileName string, dest any) (err error) {
	var (
		content []byte
	)
	filePath := getConfFile(fileName + ".json")
	if content, err = ioutil.ReadFile(filePath); err != nil {
		fmt.Println(err)
	}
	print(string(content))
	if err = json.Unmarshal(content, &dest); err != nil {
		fmt.Println(err)
	}
	return
}
