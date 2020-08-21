package HuFu

import (
	"net/url"
	"reflect"
)

type HuFu interface {
	BuildUrl()
	BuildBody(param string)
	Init()
	GetUrl() string
	GetPath() string
	GetMethod() string
	GetBody() 	string
	GetQuery()	QueryParam
}

var objMap = map[string]HuFu {
	"popOrderSearch":&PopOrderSearch{},
}
const AppKey = "87ABC15E43C1A403869411EB1432B5F5"
const AppSecret = "640B5925B240216CB87B8F0D2571E750"
const Host = "hufu.cn-north-1.jdcloud-api.net"
const Protocol = "https"
const CustomerId = "20001"

func New(api string) HuFu {
	objMap[api].Init()
	return objMap[api]
}


func NewQueryParam(method string) *QueryParam {
	return &QueryParam{
		method:method,
		app_key:AppKey,
		sign_method:"md5",
		customerId:CustomerId}
}

func CreateStrParam(param QueryParam) string {
	rt := reflect.TypeOf(param)
	rv := reflect.ValueOf(param)

	var strParam = ""

	for i:=0;i<rv.NumField();i++ {
		if strParam != "" {
			strParam += "&"
		}
		strParam += rt.Field(i).Name+"="+rv.Field(i).String()
	}

	return strParam
}

func CreateQuery(param QueryParam) url.Values {
	var query = make(url.Values)
	rt := reflect.TypeOf(param)
	rv := reflect.ValueOf(param)

	for i:=0;i<rv.NumField();i++ {
		query[rt.Field(i).Name] = []string{rv.Field(i).String()}
	}

	return query
}
