package main

import (
	"flag"
	"fmt"
	"github.com/onmpw/JYGO/config"
	"github.com/onmpw/JYGO/model"
	"orderServer/include"
	"orderServer/platform/Pdd"
	"orderServer/platform/Youzan"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	pid         = os.Getpid()
	stop        = false
	childFinish = false
)

func main() {
	var tree = map[string][3]*include.Data{
		"pdd": {
			{
				Platform:    "pdd",
				OrderStatus: "WAIT_SELLER_SEND",
				OrderInfo: &Pdd.OrderInfo{
					SyncTime: make(map[int]string),
					AddOrUp:  make(map[int]bool),
					SidToCid: make(map[int]int),
				},
			}, {
				Platform:    "pdd",
				OrderStatus: "WAIT_BUYER_CONFIRM",
				OrderInfo: &Pdd.OrderInfo{
					SyncTime: make(map[int]string),
					AddOrUp:  make(map[int]bool),
					SidToCid: make(map[int]int),
				},
			}, {
				Platform:    "pdd",
				OrderStatus: "TRADE_SUCCESS",
				OrderInfo: &Pdd.OrderInfo{
					SyncTime: make(map[int]string),
					AddOrUp:  make(map[int]bool),
					SidToCid: make(map[int]int),
				},
			}},
	}

	handleSignal()

	ModelInit()

	for {
		childFinish = false
		var shopList []*include.ShopInfo
		now := time.Unix(time.Now().Unix(), 0).Format(include.DateTimeFormat)

		num, _ := model.Read(new(include.ShopInfo)).Filter("is_delete", 0).Filter("end_date", ">", now).GetAll(&shopList)

		if num > 0 {
			include.ShopList = shopList
			for _, val := range tree {

				go start(val[0])
				go start(val[1])
				go start(val[2])
			}
		}
		wait()
	}

}

func start(v *include.Data) {
	defer func(){
		if e := recover(); e != nil {
			fmt.Printf(v.Platform+"平台"+v.OrderStatus+"订单同步 ERROR:%v\n",e)
			include.C <- 1
		}
	}()
	err := v.OrderInfo.BuildData(v.OrderStatus)

	if err != nil {
		fmt.Println(err)
	} else {
		_ = v.OrderInfo.Send()
	}

	include.C <- 1
}

func ModelInit() {
	var iniFile = flag.String("ini", "hello", "string类型参数")
	flag.Parse()
	_ = config.Init(*iniFile)
	model.Init()
	model.RegisterModel(new(Pdd.OrderTrade), new(include.ShopInfo), new(include.OrderThirdSyncTime),new(Youzan.OrderTrade))
}

func wait() {
	for i := 0; i < include.TypeNum; i++ {
		<-include.C
	}

	childFinish = true

	waitStop()

	<-time.After(time.Minute * 2)

	waitStop()
}

func testFunc() {
	for i := 0; i < 30; i++ {
		fmt.Println(i)
		<-time.After(time.Second * 1)
	}

	include.C <- 1
}

func handleSignal() {
	//signal.Ignore(os.Interrupt)
	//signal.Ignore(syscall.SIGHUP)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGTERM)
		<-c
		stop = true
		if childFinish {
			killProcess()
		}
	}()
}

func waitStop() {
	if stop {
		killProcess()
	}
}

func killProcess() {
	processor, err := os.FindProcess(pid)

	if err != nil {
		fmt.Println(err)
		return
	}

	_ = processor.Kill()
}
