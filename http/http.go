package http

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	core "git.jd.com/jcloud-api-gateway/jdcloud-apigateway-signature-go"
	"github.com/gofrs/uuid"
	"github.com/onmpw/JYGO/config"
	"io"
	"io/ioutil"
	"net/http"
	"orderServer/HuFu"
	"orderServer/include"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	SendData map[string]string
)

const (
	AppSecret = "3270064d0afe3be41ae838cd9e667b1c"
	AppId     = "1001"
)
type Jsons struct {
	order []orderInfo
	count int
}
type orderInfo struct {
	cid,sid,id int
	price	float32
	response,oid,created,modified string
}

func buildPostData() map[string]string{
	return map[string]string{
		"head":"",
		"app_id":AppId,
		"nonce": Md5(strconv.FormatInt(time.Now().Unix(),10)),
		"ip":"",
		"method":"Provider\\SyncOrderService@orderSyncDistribute",
	}
}

func setPostData(SendData *map[string]string,key string, val string) {
	(*SendData)[key] = val
}

func Exec(value string) bool {
	sendData := buildPostData()
	setPostData(&sendData,"data",value)
	sign := createSign(sendData)
	setPostData(&sendData,"sign",sign)
	res,err := post(sendData,config.Conf.C("api_host"),false)

	if err != nil {
		return false
	}

	result := parseResult(res)

	if result["errno"] != "0" {
		fmt.Println(include.Now()+":"+result["errmsg"])
		return false
	}

	return true

}

func HuFuGet(param string,api string)(interface{},error) {
	obj := HuFu.New(api)
	obj.BuildUrl()
	obj.BuildBody(param)

	// 创建带验证签名的虎符请求
	req := createHuFuRequest(obj)

	res,err := (&http.Client{Timeout:30*time.Second}).Do(req)

	if err != nil {
		fmt.Printf("虎符请求异常%v\n",err)
		return nil,err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("读取结果异常%v\n",err)
		return nil,err
	}
	var result = make(map[string]interface{})
	err = json.Unmarshal(body,&result)

	if err != nil {
		fmt.Printf("解析异常%v\n",err)
		return nil,err
	}

	if !ParseHuFuResult(result) {
		fmt.Printf("服务返回结果异常: Code=%v,Msg=%v\n",result["code"],result["msg"])
		return nil,fmt.Errorf("服务返回结果异常: Code=%v,Msg=%v\n",result["code"],result["msg"])
	}

	return decryptor(result["ext"].(string)),nil
}

// createHuFuRequest  创建带签名的虎符请求
func createHuFuRequest(obj HuFu.HuFu) *http.Request {
	req , _ := http.NewRequest("POST",obj.GetUrl(),strings.NewReader(obj.GetBody()))
	// 创建虎符头部签名
	nonce , _ := uuid.NewV4()
	req.Header.Set(core.HeaderJdcloudNonce,nonce.String())
	time := time.Now()
	formattedTime := time.UTC().Format(core.TimeFormat)
	req.Header.Set(core.HeaderJdcloudDate,formattedTime)
	req.Header.Set("x-jdcloud-algorithm","JDCLOUD2-HMAC-SHA256")
	req.Header.Set("x-jdcloud-Content-Sha256","UNSIGNED-PAYLOAD")
	req.Header.Set("Content-Type","application/json")

	Credential := *core.NewCredential(HuFu.AppKey,HuFu.AppSecret)
	Logger := core.NewDefaultLogger(3)
	signer := core.NewSigner(Credential,Logger)
	sign,_ := signer.Sign(req.Host,obj.GetPath(),"POST",req.Header,HuFu.CreateQuery(obj.GetQuery()),obj.GetBody())

	req.Header.Set(core.HeaderJdcloudAuthorization,sign)

	return req
}

// ParseHuFuResult 解析虎符返回的结果是否正常
func ParseHuFuResult(result map[string]interface{}) bool {
	if result["code"] != "0000" {
		return false
	}
	return true
}

func Get(param string,method string,host string) (interface{}, error) {
	sendData := buildPostData()
	setPostData(&sendData,"data",param)
	setPostData(&sendData,"method",method)
	sign := createSign(sendData)
	setPostData(&sendData,"sign",sign)

	return get(sendData,host,true)
}

func parseResult(value interface{}) (res map[string]string) {
	rv := reflect.ValueOf(value)

	iter := rv.MapRange()

	res = make(map[string]string)

	for iter.Next() {
		key := iter.Key()
		val := iter.Value()

		str := fmt.Sprintf("%v",val)
		res[key.String()] = str
	}

	return res
}

func post(SendData map[string]string,host string,decrypt bool) (interface{},error) {
	jsons , _ := json.Marshal(SendData)
	requestBody := string(jsons)
	res, err := http.Post(host, "application/json;charset=utf-8", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return nil,err
	}

	body, err := ioutil.ReadAll(res.Body)

	var result = make(map[string]string)
	err = json.Unmarshal(body,&result)

	if err != nil {
		return nil,err
	}
	return result,nil
}

func get(SendData map[string]string,host string,decrypt bool) (interface{},error) {
	jsons , _ := json.Marshal(SendData)
	requestBody := string(jsons)
	res, err := http.Post(host, "application/json;charset=utf-8", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return nil,err
	}

	body, err := ioutil.ReadAll(res.Body)


	if decrypt {
		r := decryptor(string(body))
		return r,nil
	}

	var result = make(map[string]interface{})
	err = json.Unmarshal(body,&result)

	if err != nil {
		return nil,err
	}
	return result["order"],nil
}

func decryptor(data string) interface{} {
	sendData := buildPostData()
	setPostData(&sendData,"data",data)
	setPostData(&sendData,"method","Provider\\DecryptService@decrypt")
	sign := createSign(sendData)
	setPostData(&sendData,"sign",sign)
	res,err := get(sendData,config.Conf.C("api_host"),false)
	if err != nil {
		return nil
	}

	return res
}

func createSign(SendData map[string]string)(sign  string){
	var contact []string
	for key,val := range SendData {
		contact = append(contact,key+val)
	}
	sort.Strings(contact)
	sign = AppSecret
	for _,str := range contact {
		sign += str
	}
	sign += AppSecret

	return strings.ToUpper(Md5(sign))
}


func Md5(value string) string {
	w := md5.New()
	_,err := io.WriteString(w,value)

	if err != nil {
		return "error"
	}

	return fmt.Sprintf("%x",w.Sum(nil))
}
