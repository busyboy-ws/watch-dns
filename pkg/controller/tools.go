package controller

import (
	"watch-dns/pkg/dns"
	aliyundns "watch-dns/pkg/dns/aliyun"
)

// cloudFlare domain
var cfDomainList =[...]string{"tuiwen-tech.com", "novelpdfdownload.com", "mickread.com","funstory.ai", "babelnovel.com"}

// maybe have other option
func SelectClient(domain string) dns.Dns  {
	return aliyundns.NewAliyunApi()
}