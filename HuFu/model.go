package HuFu

type popOrderSearchBody struct {
	OrderState,StartDate,EndDate string
	Cid,Sid	int
}

type QueryParam struct {
	method     string
	app_key     string
	sign_method string
	customerId	string
}

type PopOrderSearch struct {
	Param 			*QueryParam
	QueryBody		string
	BodyIndexName	string
	Path 			string
	Url 			string
	Method 			string
}


