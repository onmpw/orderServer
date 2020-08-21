package HuFu

import "encoding/json"

func (o *PopOrderSearch) BuildUrl() {
	url := Protocol + "://" + Host + o.Path
	strParam := CreateStrParam(*o.Param)
	url += "?"+strParam

	o.Url = url
}

func (o *PopOrderSearch) BuildBody(param string) {
	queryBody := map[string]string {
		o.BodyIndexName:param,
	}
	jsons , _ := json.Marshal(queryBody)
	o.QueryBody = string(jsons)
}


func (o *PopOrderSearch)Init() {
	o.Method = "jingdong.hufu.order.popOrderSearch"
	o.Param = NewQueryParam(o.Method)
	o.Path = "/order/getUnSensitiveData/popOrderSearch"
	o.BodyIndexName = "jingdongHufuOrderPopOrderSearchBody"
}

func (o *PopOrderSearch)GetUrl() string {
	return o.Url
}

func (o *PopOrderSearch)GetPath() string {
	return o.Path
}

func (o *PopOrderSearch)GetMethod() string {
	return o.Method
}

func (o *PopOrderSearch)GetBody() string {
	return o.QueryBody
}

func (o *PopOrderSearch)GetQuery() QueryParam {
	return *o.Param
}
