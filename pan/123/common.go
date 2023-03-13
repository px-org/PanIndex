package _123

import (
	"errors"
	jsoniter "github.com/json-iterator/go"
	"github.com/px-org/PanIndex/module"
	"github.com/px-org/PanIndex/pan/base"
	"github.com/px-org/PanIndex/util"
	"net/http"
)

func (p *Pan123) request(account *module.Account, url string, method string, callback base.Callback, resp interface{}) ([]byte, error) {
	session := Sessions[account.Id]
	req := base.Client.R().
		SetAuthToken(session.Data.Token)
	req.SetHeaders(map[string]string{
		"origin":       "https://www.123pan.com",
		"content-type": "application/json;charset=UTF-8",
		"platform":     "web",
		"app-version":  "1.2",
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
	body := res.Body()
	code := jsoniter.Get(body, "code").ToInt()
	if code != 0 {
		if code == http.StatusUnauthorized {
			_, err := p.AuthLogin(account)
			if err != nil {
				return nil, err
			}
			return p.request(account, url, method, callback, resp)
		}
		return nil, errors.New(jsoniter.Get(body, "message").ToString())
	}
	return body, nil
}

type BlockFile struct {
	Content         []byte
	partNumberStart int
	partNumberEnd   int
}

func ReadBlock(fileChunkSize int, file *module.UploadInfo) []BlockFile {
	var bfs []BlockFile
	cbs := util.ChunkBytes(file.Content, fileChunkSize)
	for i, cb := range cbs {
		partNumberStart := i + 1
		partNumberEnd := partNumberStart + 1
		bfs = append(bfs, BlockFile{cb, partNumberStart, partNumberEnd})
	}
	return bfs
}
