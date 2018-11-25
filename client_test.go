package httpfsclient_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/RocksonZeta/httpfsclient"
	"github.com/stretchr/testify/assert"
)

const redisAddr = "localhost:6379"
const clusterId = "static"

//eg hflink: "static:s1/txt/0/0/qlkgmiidid/h9ahc8auc7.txt"
func TestBasic(t *testing.T) {
	httpfsclient.InitClusters(redisAddr, "", "0", clusterId)
	file := "client.go"
	info, err := os.Stat(file)
	assert.Nil(t, err)
	buff, _ := os.Open("client.go")
	defer buff.Close()
	link, err := httpfsclient.Write(buff, clusterId, "client.go", httpfsclient.CollectionTxt)
	fmt.Println(link)
	assert.Nil(t, err)
	stat, err := link.Stat()
	assert.Nil(t, err)
	assert.Equal(t, info.Name(), stat.RawName)
	bs, err := link.Read()
	assert.Nil(t, err)
	assert.Equal(t, len(bs), int(stat.Size))
}
func TestVideo(t *testing.T) {
	httpfsclient.InitClusters(redisAddr, "", "0", clusterId)
	link := httpfsclient.HfLink("static:s1/video/0/0/9o39m9wuvi/4uie3br1wj.mp4")
	err := link.VideoCompressDash("v1/progress")
	assert.Nil(t, err)
}
func TestImage(t *testing.T) {
	httpfsclient.InitClusters(redisAddr, "", "0", clusterId)
	link := httpfsclient.HfLink("static:s1/image/0/0/gysz2c6aqf/joexrtxyco.jpg")
	links, err := link.ImageResize([]int{10, 10, 100, 100}, [][]int{{60, 60}})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(links))
}
