package Jd

import (
	"encoding/json"
	"fmt"
	"github.com/onmpw/JYGO/model"
	"orderServer/http"
	"orderServer/include"
	"reflect"
	"strconv"
	"strings"
)

var OrderStatus = map[string]string {
	"WAIT_SELLER_SEND":"WAIT_SELLER_STOCK_OUT",
	"WAIT_BUYER_CONFIRM":"WAIT_GOODS_RECEIVE_CONFIRM",
	"TRADE_SUCCESS":"FINISHED_L",
	"TRADE_CANCEL":"LOCKED,TRADE_CANCELED",
}
var platform = "J"

func (o *OrderInfo) BuildData(orderStatus string) error{
	var start string
	var end string
	var flag bool
	o.orderStatus = orderStatus
	o.order = o.order[0:0]

	//var trades []*OrderTrade
	//getParam := map[string]string{
	//	"order_state":OrderStatus[orderStatus],
	//	"cid": "143881",
	//	"sid": "816287",
	//	"start_date": "2020-04-02 01:58:51",
	//	"end_date":"2020-04-02 15:58:51",
	//}
	//getParam := map[string]string{
	//	"type":OrderStatus[orderStatus],
	//	"cid": "143881",
	//	"sid": "816287",
	//	"startDate": "2020-04-02 01:58:51",
	//	"endDate":"2020-04-02 15:58:51",
	//}
	//num,_ := getOrder(getParam,&trades)
	//fmt.Println(trades)
	//fmt.Println(num)
	for _,shop := range include.ShopList {
		var trades []*OrderTrade
		if shop.Type != platform {
			continue
		}
		var t *include.OrderThirdSyncTime
		count := model.Read(new(include.OrderThirdSyncTime)).Filter("platform",platform).Filter("type",OrderStatus[orderStatus]).Filter("company_id",shop.Cid).Filter("sid",shop.Sid).Count()

		if count >= 1 {
			err := model.Read(new(include.OrderThirdSyncTime)).Filter("platform", platform).Filter("type", OrderStatus[orderStatus]).Filter("company_id", shop.Cid).Filter("sid", shop.Sid).GetOne(&t)
			if err != nil {
				return err
			}
			start = t.Created
			flag = true    // true 表示记录存在 需要更新
		}else {
			start = include.GetNewShopTime()
			flag = false  // false 表示记录不存在 需要新增
		}
		end = include.Now()

		getParam := map[string]string{
			"order_state":OrderStatus[orderStatus],
			"cid": strconv.Itoa(shop.Cid),
			"sid": strconv.Itoa(shop.Sid),
			"start_date": start,
			"end_date":end,
		}
		num,_ := getOrder(getParam,&trades)
		// 获取订单
		//num , _ := model.Read(new(OrderTrade)).Filter("type", OrderStatus[orderStatus]).Filter("cid", shop.Cid).Filter("sid", shop.Sid).Filter("modified",">=",start).Filter("modified","<",end).GetAll(&trades)
		o.SyncTime[shop.Sid] = start
		o.AddOrUp[shop.Sid] = flag
		o.SidToCid[shop.Sid] = shop.Cid
		if num <= 0 {
			continue
		}

		o.order = append(o.order,trades...)
		o.getMaxTime(trades,shop.Sid)
	}

	return nil
}

func getOrder(param map[string]string,models *[]*OrderTrade)(num int,err error ){
	jsons, err := json.Marshal(param)
	if err != nil {
		fmt.Printf("京东订单同步错误：%v",err)
		return 0,err
	}

	//trades,_ := http.Get(string(jsons),"Provider\\OrderService@search",config.Conf.C("jd_api_host"))
	// 改造为使用虎符方式
	trades,_ := http.HuFuGet(string(jsons),"popOrderSearch")

	*models = parseOrderList(trades)

	return len(*models),err
}

func parseOrderList(value interface{}) (orderList []*OrderTrade) {
	rv := reflect.ValueOf(value)

	orderList = make([]*OrderTrade,rv.Len())
	if rv.Kind() == reflect.Slice {
		l := rv.Len()

		for i:=0;i<l;i++ {
			iv := reflect.Indirect(rv.Index(i))

			r := iv.Interface().(map[string]interface{})

			var trade = new(OrderTrade)
			trade.Id = int(r["id"].(float64))
			trade.Cid = int(r["cid"].(float64))
			trade.Sid = int(r["sid"].(float64))
			trade.Oid = r["oid"].(string)
			trade.Response = r["response"].(string)
			trade.Created = r["created"].(string)
			trade.Modified = r["modified"].(string)
			trade.Type = r["type"].(string)
			orderList[i] = trade
		}
	}
	return orderList
}

func (o *OrderInfo) Send() bool {
	var order string
	if len(o.order) > 0{
		jsons, err := json.Marshal(o.order)

		if err != nil {
			return false
		}

		order  = string(jsons)
	}

	data := map[string]string {
		"platform":"jd",
		"order_status":o.orderStatus,
		"order_list":order,
	}

	jsons, err := json.Marshal(data)
	if err != nil {
		return false
	}
	o.updateSyncTime()
	return http.Exec(string(jsons))
}

func (o *OrderInfo) getMaxTime(trades []*OrderTrade,sid int) {
	if len(trades) <= 0 {
		return
	}

	for _,trade := range trades {
		if strings.Compare(trade.Modified,o.SyncTime[sid]) == 1 {
			o.SyncTime[sid] = trade.Modified
		}
	}
}

func (o *OrderInfo) updateSyncTime() {
	var syncTime include.OrderThirdSyncTime

	syncTime.Type = OrderStatus[o.orderStatus]
	syncTime.Platform = platform
	syncTime.Updatetime = include.Now()
	for sid,created := range o.SyncTime {
		syncTime.Sid = sid
		syncTime.Created = created
		syncTime.Company_id = o.SidToCid[sid]
		if o.AddOrUp[sid] { // 需要更新
			where := []interface{}{[]interface{}{"company_id",o.SidToCid[sid]},[]interface{}{"platform",platform},[]interface{}{"sid",sid},[]interface{}{"type",OrderStatus[o.orderStatus]}}
			_ , err := model.Update(syncTime,where)
			if err != nil {
				fmt.Println(err)
			}
		}else {
			_,err := model.Add(syncTime)
			if err != nil {
				fmt.Println(err)
			}
		}
	}
}
