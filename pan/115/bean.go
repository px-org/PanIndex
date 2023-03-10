package _115

type UploadResp struct {
	Object    string `json:"object"`
	Accessid  string `json:"accessid"`
	Host      string `json:"host"`
	Policy    string `json:"policy"`
	Signature string `json:"signature"`
	Expire    int    `json:"expire"`
	Callback  string `json:"callback"`
}
