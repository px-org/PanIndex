package entity

type FileNode struct {
	FileId       string `json:"fileId" gorm:"primary_key"`
	FileIdDigest string `json:"fileIdDigest"`
	FileName     string `json:"fileName"`
	FileSize     int64  `json:"fileSize"`
	SizeFmt      string `json:"sizeFmt"`
	FileType     string `json:"fileType"`
	IsFolder     bool   `json:"isFolder"`
	IsStarred    bool   `json:"isStarred"`
	LastOpTime   string `json:"lastOpTime"`
	ParentId     string `json:"parentId"`
	Path         string `json:"path"`
	ParentPath   string `json:"parentPath"`
	DownloadUrl  string `json:"downloadUrl"`
	MediaType    int    `json:"mediaType"` //1图片，2音频，3视频，4文本文档，0其他类型
	Icon         Icon   `json:"icon"`
	CreateTime   string `json:"create_time"`
	Delete       int    `json:"delete"`
	Hide         int    `json:"hide"`
}
type Icon struct {
	LargeUrl string `json:"largeUrl"`
	SmallUrl string `json:"smallUrl"`
}
type Paths struct {
	FileId    string `json:"fileId"`
	FileName  string `json:"fileName"`
	IsCoShare int    `json:"isCoShare"`
}
