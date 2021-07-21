package aliyun

import (
	utils "watch-dns/pkg/tools"
	dns "watch-dns/pkg/dns"
	alidns "github.com/alibabacloud-go/alidns-20150109/v2/client"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/emicklei/go-restful/log"
	"strings"
	openapi "github.com/alibabacloud-go/darabonba-openapi/client"
)



type aliyunDns struct {
	client *alidns.Client
}

func NewAliyunApi() dns.Dns {
	aliyunKey, keySecret := utils.GetAliyunKey()
	config := &openapi.Config{AccessKeyId: aliyunKey,AccessKeySecret: keySecret}
	config.Endpoint = tea.String("dns.aliyuncs.com")
	result, err := alidns.NewClient(config)
	if err != nil{
		panic(err)
	}
	cli := &aliyunDns{client: result}
	return cli

}

func (client *aliyunDns)QueryDns(rr string, domain string) (string, bool)  {
	describeDomainRecordsRequest := &alidns.DescribeDomainRecordsRequest{
		DomainName: tea.String(domain),
	}
	result, err := client.client.DescribeDomainRecords(describeDomainRecordsRequest)
	if err != nil{panic(err)}
	for _, val := range result.Body.DomainRecords.Record {
		if strings.EqualFold(*val.RR, rr ){
			return *val.RecordId, true
		}
	}
	return "", false
}

func (client *aliyunDns)UpdateDns(rr string, domain string, ip string) bool  {
	recordID, flag := client.QueryDns(rr, domain)
	if !flag{
		return false
	}
	updateDomainRecordRequest := &alidns.UpdateDomainRecordRequest{
		RecordId: tea.String(recordID),
		RR: tea.String(rr),
		Type: tea.String("A"),
		Value: tea.String(ip),
	}
	_, err := client.client.UpdateDomainRecord(updateDomainRecordRequest)
	if err !=nil {
		log.Printf("request error: ", err)
		return false
	}
	return true
}

func (client *aliyunDns)AddDns(rr string, ipadddr string, domain string) bool  {
	addDomainRecordRequest := &alidns.AddDomainRecordRequest{
		DomainName: tea.String(domain),
		Type: tea.String("A"),
		Value: tea.String(ipadddr),
		RR: tea.String(rr),
	}
	_, err := client.client.AddDomainRecord(addDomainRecordRequest)
	if err != nil{
		log.Printf("request error: ", err)
		return false
	}
	return true
}
