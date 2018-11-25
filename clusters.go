package httpfsclient

import (
	"sync"
	"time"

	"github.com/httpfsclient/kv"
)

var clusters *Clusters
var serverUts sync.Map

func init() {
	serverUts = sync.Map{} //clusterId:serverId => ServerUt
	clusters = new(Clusters)
}

type ServerUt struct {
	Ut          int64
	UpdateCount int
}

func GetClusters() *Clusters {
	return clusters
}

func GetServer(clusterId, serverId string) Server {
	cs := GetClusters()
	if nil == cs {
		return Server{}
	}
	return cs.GetServer(clusterId, serverId)
}

func InitClusters(addr, password, db string, clusterIds ...string) {
	conf := kv.RedisConfig{
		Addr:     addr,
		Password: password,
		Database: db,
	}
	factory := kv.NewFactory(&conf)
	load(factory, clusterIds...)
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		for {
			select {
			case <-ticker.C:
				load(factory, clusterIds...)
			}
		}
	}()
}

// load redis clusters
func load(factory *kv.ServiceFactory, clusterIds ...string) {
	redis := factory.Get()
	defer redis.Close()
	for _, cid := range clusterIds {
		var servers map[string]Server
		redis.HMGetAll(cid, &servers)
		cluster := Cluster{Id: cid}
		for k, v := range servers {
			v.available = available(v)
			cluster.servers.Store(k, v)
		}
		clusters.clusters.Store(cid, cluster)
	}
}
func available(server Server) bool {
	if s, ok := serverUts.Load(server.Id()); ok {
		if su, ok1 := s.(ServerUt); ok1 {
			var r bool
			if su.Ut != server.Ut {
				r = true
				su.UpdateCount = 0
			} else {
				r = su.UpdateCount <= 5
				su.UpdateCount++
			}
			su.Ut = server.Ut
			serverUts.Store(server.Id(), su)
			return r
		}
	}
	serverUts.Store(server.Id(), ServerUt{Ut: server.Ut, UpdateCount: 0})
	return true
}

//Clusters include many clusters
type Clusters struct {
	// clusters map[string]Cluster // clusterId :Server
	clusters sync.Map // clusterId :*Server
}

func (c *Clusters) GetCluster(clusterId string) (Cluster, bool) {
	if cluster, ok := c.clusters.Load(clusterId); ok {
		return cluster.(Cluster), ok
	}
	return Cluster{}, false
}
func (c *Clusters) GetServer(clusterId, serverId string) Server {
	if cluster, ok := c.clusters.Load(clusterId); ok {
		return cluster.(Cluster).GetServer(serverId)
	}
	return Server{}
}
func (c *Clusters) Url(clusterId, serverId string) string {
	if cluster, ok := c.clusters.Load(clusterId); ok {
		return cluster.(Cluster).Url(serverId)
	}
	return ""
}
func (c *Clusters) HfsId(url string) (clusterId, serverId string) {
	c.clusters.Range(func(k, v interface{}) bool {
		clusterId, serverId = v.(*Clusters).HfsId(url)
		if serverId != "" {
			return false
		}
		return true
	})
	return
}

type Cluster struct {
	Id string
	// serverm map[string]Server //serverId : Server
	servers sync.Map //serverId : Server
}

func (c Cluster) ChooseServer() Server {
	var r Server
	var maxFreeSpace int
	c.servers.Range(func(k, v interface{}) bool {
		s := v.(Server)
		if s.FreeSpace > maxFreeSpace && s.available {
			maxFreeSpace = s.FreeSpace
			r = s
		}
		return true
	})
	return r
}

func (c Cluster) Url(serverId string) string {
	if v, ok := c.servers.Load(serverId); ok {
		return v.(Server).Proxy
	}
	return ""
}
func (c Cluster) HfsId(url string) (clusterId, serverId string) {
	c.servers.Range(func(k, v interface{}) bool {
		s := v.(Server)
		if s.Proxy == url {
			clusterId = s.ClusterId
			serverId = s.ServerId
			return false
		}
		return true
	})
	return
}
func (c Cluster) GetServer(serverId string) Server {
	if v, ok := c.servers.Load(serverId); ok {
		return v.(Server)
	}
	return Server{}
}
