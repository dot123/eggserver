package schema

type LoginReq struct {
	UserUid      string `json:"userUid" msgpack:"userUid" binding:"required"`
	UserName     string `json:"username" msgpack:"username"`
	FirstName    string `json:"firstName" msgpack:"firstName"`
	LastName     string `json:"lastName" msgpack:"lastName"`
	PhotoUrl     string `json:"photoUrl" msgpack:"photoUrl"`
	LanguageCode string `json:"languageCode" msgpack:"languageCode"`
	OS           string `json:"os" msgpack:"os"`
	StartParam   string `json:"startParam" msgpack:"startParam"`
}

type LoginResp struct {
	Token string `json:"token" msgpack:"token"`
}
