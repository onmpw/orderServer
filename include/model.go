package include


type ShopInfo struct {
	Sid		int
	Cid		int
	Type 	string
}

func (s *ShopInfo) TableName() string {
	return "shop_taobao"
}


type OrderThirdSyncTime struct {
	Type,Platform,Updatetime 	string
	Sid, Company_id	int
	Created	string
}

func (o *OrderThirdSyncTime) TableName() string {
	return "order_thirdsync_time"
}