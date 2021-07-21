package utils

import (
	"github.com/alibabacloud-go/tea/tea"
	"os"
)
func GetClusterDomain() string {
	return os.Getenv("CLUSTER_DOMAIN")
}

func GetAliyunKey()(*string,*string)  {
	accessKeyId := os.Getenv("ALIYUN_KEYID")
	accessKeySecret := os.Getenv("ALIYUN_KEYSECRET")
	if accessKeyId == "" || accessKeySecret == ""{
		panic("get aliyun key error")
	}
	return tea.String(accessKeyId), tea.String(accessKeySecret)
}

func GetCfApiKey() string {
	apiKey := os.Getenv("CF_APIKEY")
	if apiKey == ""{
		panic("get aliyun key error")
	}
	return apiKey
}