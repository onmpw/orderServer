package http

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/onmpw/JYGO/config"
	"io"
	"io/ioutil"
	"net/http"
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

func buildPostData() {
	SendData = map[string]string{
		"head":"",
		"app_id":AppId,
		"nonce": Md5(strconv.FormatInt(time.Now().Unix(),10)),
		"ip":"",
		"method":"Provider\\SyncOrderService@orderSyncDistribute",
	}
}

func setPostData(key string, val string) {
	SendData[key] = val
}

func Exec(value string) bool {
	buildPostData()
	setPostData("data",value)
	sign := createSign()
	setPostData("sign",sign)

	res,err := post()

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

func post() (interface{},error) {
	jsons , _ := json.Marshal(SendData)
	requestBody := string(jsons)
	res, err := http.Post(config.Conf.C("api_host"), "application/json;charset=utf-8", bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return nil,err
	}

	body, err := ioutil.ReadAll(res.Body)

	var result = make(map[string]interface{})
	err = json.Unmarshal(body,&result)

	if err != nil {
		return nil,err
	}

	return result,nil
}

func createSign()(sign  string){
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
