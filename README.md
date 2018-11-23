# httpfsclient
httpfs client.


# Example:
```go
const redisAddr = "localhost:6379"
const clusterId = "static"

func init() {
	httpfsclient.InitClusters(redisAddr, "", "0", clusterId)
}

func WriteAndRead(file string) {
	buff, _ := os.Open(file)
	defer buff.Close()
	//write file to cluster
	link, _ := httpfsclient.Write(buff, clusterId, file, CollectionTxt)
	fmt.Println(link.Url())
	//read file
	bs, _ := link.Read()
	fmt.Println(string(bs))
	// crop and resize image
	links, err := link.ImageResize([]int{10, 10, 100, 100}, [][]int{{60, 60}})
}


```


# Dependency
```
github.com/mozillazg/request
github.com/gomodule/redigo/redis
```