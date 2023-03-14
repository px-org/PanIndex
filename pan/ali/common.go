package ali

import (
	"github.com/bluele/gcache"
)

var Alis = map[string]TokenResp{}
var APPID = "5dde4e1bdf9e4966b387ba58f4b3fdc3"
var NonceMin = 0
var NonceMax = 2147483647
var SignCache = gcache.New(100000).LRU().Build()
