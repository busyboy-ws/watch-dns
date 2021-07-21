package aliyun

import (
	"net"
	"testing"
)

func TestAliyun(T *testing.T)  {
	aliyun := NewAliyunApi()
	aliyun.AddDns("test-go","8.8.8.8","mulanread.net")

	// Query

	_, flag := aliyun.QueryDns("test-go","mulanread.net")
	if !flag{
		T.Errorf("should not query dns record ?")
	}
	host := "test-go.mulanread.net"
	ipaddr, err := net.LookupHost(host)
	if err != nil{
		T.Error(err)
	}
	if ipaddr[0] != "8.8.8.8"{
		T.Errorf("dns record many be error")
	}

}
