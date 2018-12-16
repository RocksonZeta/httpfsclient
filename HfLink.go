package httpfsclient

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"reflect"
	"strings"
)

// HfLink := clusterId:serverId/relativePath
type HfLink string

func IsHfLink(url string) bool {
	i := strings.Index(url, "/")
	if i == -1 {
		return false
	}
	head := url[0:i]
	return strings.Index(head, ":") != -1
}

func NewHfLink(clusterId, serverId, filepath string) HfLink {
	return HfLink(path.Join(clusterId+":"+serverId, filepath))
}
func urlPathStartIndex(url string) int {
	pi := strings.Index(url, "://")
	if -1 == pi {
		return strings.Index(url, "/")
	}
	i := strings.Index(url[pi+3:], "/")
	if -1 == i {
		return -1
	}
	return i + 3 + pi
}

// http://xxx/image/1.jpg -> s:1/image/1.jpg
func FromUrl(url string) (HfLink, bool) {
	pi := strings.Index(url, "://")
	i := strings.Index(url[pi+3:], "/") + 3 + pi
	proxy := url[0:i]
	clusterId, serverId := GetClusters().HfsId(proxy)
	if clusterId == "" || serverId == "" {
		return HfLink(url), false
	}

	return NewHfLink(clusterId, serverId, url[i:]), true
}

// eg. s:1/txt/00/00/yyfoatapk5/bdu9kjosiq.go -> http://xxx/txt/00/00/yyfoatapk5/bdu9kjosiq.go
func (d HfLink) Url() string {
	clusterId, serverId, path := d.Parts()
	return GetClusters().GetServer(clusterId, serverId).Proxy + path
}
func (d HfLink) String() string {
	return string(d)
}

func (d HfLink) Stat() (FileInfo, error) {
	clusterId, serverId, path := d.Parts()
	server := GetServer(clusterId, serverId)
	if "" == server.ClusterId {
		return FileInfo{}, errors.New("no such server:" + string(d))
	}
	return (&Client{Server: server.Local}).Stat(path)
}
func (d HfLink) Read() ([]byte, error) {
	clusterId, serverId, path := d.Parts()
	server := GetServer(clusterId, serverId)
	if "" == server.ClusterId {
		return nil, errors.New("no such server:" + string(d))
	}
	return (&Client{Server: server.Local}).Read(path)
}
func (d HfLink) Call(module, method string, args, result interface{}) error {
	clusterId, serverId, _ := d.Parts()
	return Methods{}.Call(clusterId, serverId, module, method, args, result)
}
func (d HfLink) CallAsync(module, method string, args interface{}) error {
	clusterId, serverId, _ := d.Parts()
	return Methods{}.CallAsync(clusterId, serverId, module, method, args)
}
func (d HfLink) ImageResize(crop []int, sizes [][]int) ([]HfLink, error) {
	return Methods{}.ImageCropResize(d, crop, sizes)
}
func (d HfLink) VideoCompressDash(videoId int, redisProgressKey string) error {
	return Methods{}.VideoCompressDash(d, videoId, redisProgressKey)
}

func (d HfLink) Path() string {
	ds := string(d)
	i := strings.Index(ds, "/")
	if i == -1 || '/' == ds[0] {
		return ds
	}
	return ds[i:]
}
func (d HfLink) Parts() (string, string, string) {
	ds := string(d)
	i := strings.Index(ds, "/")
	if i == -1 || '/' == ds[0] {
		return "", "", ds
	}
	server := ds[0:i]
	j := strings.Index(server, ":")
	if j == -1 {
		return "", "", ds
	}
	return server[0:j], server[j+1:], ds[i:]
}

type NullHfLink struct{ sql.NullString }

// NewString creates a new String
func NewNullHfLink(s HfLink, valid bool) NullHfLink {
	return NullHfLink{
		NullString: sql.NullString{
			String: string(s),
			Valid:  valid,
		},
	}
}
func FromHfLink(s HfLink) NullHfLink {
	return NullHfLink{
		NullString: sql.NullString{
			String: string(s),
			Valid:  len(s) == 0,
		},
	}
}

func (s *NullHfLink) UnmarshalJSON(data []byte) error {
	var err error
	var v interface{}
	if err = json.Unmarshal(data, &v); err != nil {
		return err
	}
	switch x := v.(type) {
	case string:
		if IsHfLink(x) {
			s.String = x
		} else {
			dd, _ := FromUrl(x)
			s.String = string(dd)
		}
	case map[string]interface{}:
		err = json.Unmarshal(data, &s.NullString)
	case nil:
		s.Valid = false
		return nil
	default:
		err = fmt.Errorf("json: cannot unmarshal %v into Go value of type null.String", reflect.TypeOf(v).Name())
	}
	s.Valid = err == nil
	return err
}

// MarshalJSON implements json.Marshaler.
// It will encode null if this String is null.
func (s NullHfLink) MarshalJSON() ([]byte, error) {
	if !s.Valid {
		return []byte("null"), nil
	}
	return json.Marshal(HfLink(s.String).Url())
}

func (s NullHfLink) Url() string {
	return HfLink(s.String).Url()
}

// MarshalText implements encoding.TextMarshaler.
// It will encode a blank string when this String is null.
func (s NullHfLink) MarshalText() ([]byte, error) {
	if !s.Valid {
		return []byte{}, nil
	}
	return []byte(HfLink(s.String).Url()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
// It will unmarshal to a null String if the input is a blank string.
func (s *NullHfLink) UnmarshalText(text []byte) error {
	s.String = HfLink(string(text)).Url()
	s.Valid = s.String != ""
	return nil
}

// SetValid changes this String's value and also sets it to be non-null.
func (s *NullHfLink) SetValid(v string) {
	s.String = v
	s.Valid = true
}

// Ptr returns a pointer to this String's value, or a nil pointer if this String is null.
func (s NullHfLink) Ptr() *string {
	if !s.Valid {
		return nil
	}
	return &s.String
}

// IsZero returns true for null strings, for potential future omitempty support.
func (s NullHfLink) IsZero() bool {
	return !s.Valid
}
