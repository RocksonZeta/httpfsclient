package httpfsclient

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/RocksonZeta/httpfsclient/util/httputil"
	"github.com/mozillazg/request"
)

const (
	CollectionImage  = "image"
	CollectionVideo  = "video"
	CollectionEpub   = "epub"
	CollectionTxt    = "txt"
	CollectionPdf    = "pdf"
	CollectionBin    = "bin"
	CollectionOffice = "office"
	CollectionZip    = "zip"
)

const (
	FileStateNotStart  = 0
	FileStateUploading = 1
	FileStateOk        = 2
	FileStateFail      = 3
)

type Client struct {
	Server string
}

func ParseResult(bs []byte, result interface{}) error {
	var jr JsonResult
	jr.Data = result
	err := json.Unmarshal(bs, &jr)
	return err
}

func (c *Client) Stat(filePath string) (FileInfo, error) {
	var stat FileInfo
	_, bs, err := httputil.HttpGet3(c.Server + "/fs/stat" + filePath)
	ParseResult(bs, &stat)
	return stat, err
}
func (c *Client) Ls(filePath string) ([]FileInfo, error) {
	var ls []FileInfo
	_, bs, err := httputil.HttpGet3(c.Server + "/fs/ls" + filePath)
	ParseResult(bs, &ls)
	return ls, err
}
func (c *Client) Read(filePath string) ([]byte, error) {
	// var bs []byte
	_, bs, err := httputil.HttpGet3(c.Server + "/fs/read" + filePath)
	// ParseResult(bs, &ls)
	return bs, err
}
func (c *Client) Call(module, method string, args interface{}, result interface{}) error {
	jsonArgs, err := json.Marshal(args)
	if err != nil {
		return err
	}
	status, bs, err := httputil.HttpPostForm3(c.Server+"/call/"+module+"/"+method, map[string]string{"args": string(jsonArgs)}, nil)
	if err != nil {
		return err
	}
	if status != 200 {
		return errors.New(string(bs))
	}
	return ParseResult(bs, result)
}

type ImageTransformParam struct {
	FilePath string
	Crop     []int
	Resize   [][]int
}

func (c *Client) ImageCropResize(filePath string, crop []int, sizes [][]int) (result []string, err error) {
	err = c.Call("image", "cropresize", ImageTransformParam{FilePath: filePath, Crop: crop, Resize: sizes}, &result)
	if err != nil {
		return
	}
	return
}

type Writer struct {
	ClusterId, ServerId string
}

func (w *Writer) Write(reader io.Reader, fileName, collection string) (HfLink, error) {
	server := GetClusters().GetServer(w.ClusterId, w.ServerId)
	return WriteServer(server, reader, fileName, collection)
}
func WriteServer(server Server, reader io.Reader, fileName, collection string) (HfLink, error) {
	_, bs, err := httputil.HttpPostForm3(server.Local+"/fs/write/"+collection, nil, []request.FileField{{FieldName: "file", FileName: fileName, File: reader}})
	var rpath string
	ParseResult(bs, &rpath)
	return NewHfLink(server.ClusterId, server.ServerId, rpath), err
}

func Write(reader io.Reader, clusterId, fileName, collection string) (HfLink, error) {
	cluster, ok := GetClusters().GetCluster(clusterId)
	if !ok {
		return HfLink(""), errors.New("no such cluster:" + clusterId)
	}
	server := cluster.ChooseServer()
	if "" == server.Local {
		return HfLink(""), errors.New(" cluster '" + clusterId + "' no avaiable server.")
	}
	return WriteServer(server, reader, fileName, collection)
}

type Methods struct {
}

func (c Methods) Call(clusterId, serverId, module, method string, args interface{}, result interface{}) error {
	jsonArgs, err := json.Marshal(args)
	if err != nil {
		return err
	}
	server := GetClusters().GetServer(clusterId, serverId)
	status, bs, err := httputil.HttpPostForm3(server.Local+"/call/"+module+"/"+method, map[string]string{"args": string(jsonArgs)}, nil)
	if err != nil {
		return err
	}
	if status != 200 {
		return errors.New(string(bs))
	}
	return ParseResult(bs, result)
}
func (c Methods) CallAsync(clusterId, serverId, module, method string, args interface{}) error {
	jsonArgs, err := json.Marshal(args)
	if err != nil {
		return err
	}
	server := GetClusters().GetServer(clusterId, serverId)
	status, bs, err := httputil.HttpPostForm3(server.Local+"/call/async/"+module+"/"+method, map[string]string{"args": string(jsonArgs)}, nil)
	if err != nil {
		return err
	}
	if status != 200 {
		return errors.New(string(bs))
	}
	return ParseResult(bs, nil)
}

func (m Methods) ImageCropResize(hf HfLink, crop []int, sizes [][]int) ([]HfLink, error) {
	if len(crop) != 0 && len(crop) != 4 {
		return nil, errors.New("ImageCropResize crop param error. crop must be [x,y,w,h].")
	}
	if len(sizes) > 0 {
		for _, size := range sizes {
			if len(size) != 2 {
				return nil, errors.New("ImageCropResize sizes param error. sizes must be [[w,h]].")
			}
		}
	}
	clusterId, serverId, path := hf.Parts()
	var resultPaths []string
	err := m.Call(clusterId, serverId, "image", "cropresize", ImageTransformParam{FilePath: path, Crop: crop, Resize: sizes}, &resultPaths)
	if err != nil {
		return nil, err
	}
	result := make([]HfLink, len(resultPaths))
	for i, v := range resultPaths {
		result[i] = NewHfLink(clusterId, serverId, v)
	}
	return result, nil
}

type VideoCompressParam struct {
	File             string
	ProgressRedisKey string
	VideoId          int
}

func (m Methods) VideoCompressDash(hf HfLink, progressKey string) error {
	clusterId, serverId, path := hf.Parts()
	return m.CallAsync(clusterId, serverId, "video", "dash", VideoCompressParam{File: path, ProgressRedisKey: progressKey})
}
