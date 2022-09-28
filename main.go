package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

var (
	client *dnspod.Client
	config conf_struct
)

type line_struct struct {
	Interface string
	IP        string
	SubDomain string
}
type conf_struct struct {
	ID         string
	Key        string
	Domain     string
	Interfaces []*line_struct
}
type rec_struct struct {
	ID    uint64
	Value string
}

func loadconfig() {
	b, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	json.Unmarshal(b, &config)
}

func localip(name string) string {
	eth, _ := net.InterfaceByName(name)
	addr, _ := eth.Addrs()
	for _, i := range addr {
		s := i.String()
		ss := strings.Split(s, "/")
		sa := net.ParseIP(ss[0])
		if sa.To4() != nil {
			return sa.String()
		}
	}
	return "0.0.0.0"
}
func updatelocalip() {
	for _, i := range config.Interfaces {
		if i.Interface != "" {
			i.IP = localip(i.Interface)
		}
	}
}

func main() {

	recs := getrecordlist(config.Domain)
	if recs == nil || len(recs) < 1 {
		return
	}
	for {
		updatelocalip()
		for _, i := range config.Interfaces {
			if i.IP == "0.0.0.0" {
				continue
			}
			if rec, ok := recs[i.SubDomain]; ok {
				if rec.Value == i.IP {
					continue
				}
				if update(rec.ID, i.SubDomain, config.Domain, i.IP) {
					rec.Value = i.IP
				}
			}
		}

		time.Sleep(time.Minute)
	}

}
func getrecordlist(domain string) map[string]*rec_struct {

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := dnspod.NewDescribeRecordListRequest()
	request.Domain = common.StringPtr(domain)
	// 返回的resp是一个DescribeRecordListResponse的实例，与请求对象对应
	response, err := client.DescribeRecordList(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Printf("An API error has returned: %s", err)
		return nil
	}
	if err != nil {
		log.Printf("An API error has returned: %s", err)
		return nil
	}
	out := make(map[string]*rec_struct)
	for i := range response.Response.RecordList {
		var t rec_struct
		t.ID = *response.Response.RecordList[i].RecordId
		t.Value = *response.Response.RecordList[i].Value
		out[*response.Response.RecordList[i].Name] = &t
	}
	return out
}
func init() {
	loadconfig()

	credential := common.NewCredential(
		config.ID, config.Key,
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "dnspod.tencentcloudapi.com"
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ = dnspod.NewClient(credential, "", cpf)

}
func update(recordid uint64, subdomain, domain, ip string) bool {

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := dnspod.NewModifyDynamicDNSRequest()

	request.Domain = common.StringPtr(domain)
	request.SubDomain = common.StringPtr(subdomain)
	request.RecordId = common.Uint64Ptr(recordid)
	request.RecordLine = common.StringPtr("默认")
	request.Value = common.StringPtr(ip)

	// 返回的resp是一个ModifyDynamicDNSResponse的实例，与请求对象对应
	_, err := client.ModifyDynamicDNS(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		log.Printf("An API error has returned: %s", err)
		return false
	}
	if err != nil {
		log.Printf("An API error has returned: %s", err)
		return false

	}
	return true
	// 输出json格式的字符串回包
}
