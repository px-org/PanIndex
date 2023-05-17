package _alishare

import (
	"github.com/bluele/gcache"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/base"
)

var Sessions = map[string]ShareTokenResp{}
var APPID = "5dde4e1bdf9e4966b387ba58f4b3fdc3"
var NonceMin = 0
var NonceMax = 2147483647
var SignCache = gcache.New(100000).LRU().Build()

func (a *AliShare) request(account *module.Account, url string, method string, callback base.Callback, resp interface{}) ([]byte, error) {
	session := Sessions[account.Id]
	req := base.Client.R().SetHeaders(map[string]string{
		"origin":        "https://www.aliyundrive.com",
		"content-type":  "application/json;charset=UTF-8",
		"x-share-token": session.ShareToken,
	})
	if callback != nil {
		callback(req)
	}
	if resp != nil {
		req.SetResult(resp)
	}
	res, err := req.Execute(method, url)
	if err != nil {
		return nil, err
	}
	return res.Body(), nil
}
