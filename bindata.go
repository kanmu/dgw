// Code generated by go-bindata. DO NOT EDIT.
// sources:
// template/method.tmpl (1.57kB)
// template/struct.tmpl (206B)

package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("read %q: %w", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type asset struct {
	bytes  []byte
	info   os.FileInfo
	digest [sha256.Size]byte
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _templateMethodTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xbc\x93\x4f\x8b\x9c\x40\x10\xc5\xcf\xd3\x9f\xa2\x72\x09\x1a\x3a\xed\x3d\xb0\x87\xec\x92\x2c\x81\x25\xcc\xc4\x40\x8e\xd9\xb6\x2d\xb3\xa2\xd3\x42\x59\x26\xb3\x88\xdf\x3d\xd8\xa3\xc6\x7f\x83\x7b\x98\x6c\x5f\x06\xa6\x9f\xd5\xbf\x7a\xaf\x2a\x08\xe0\x8e\x50\x33\x42\x6a\x4b\x24\x2e\x81\x9f\x10\xea\x1a\x54\xc8\x54\x19\x56\x5f\xf5\x11\xa1\x69\x80\x0b\x77\x13\x6b\xd6\x91\x2e\x51\x89\xa4\xb2\x06\x3c\x82\x77\x4b\xb1\xdf\xd5\xf4\xe2\x08\x0e\x15\xd2\x33\x92\x0f\x48\x54\x10\xd4\x02\x80\x90\x2b\xb2\x40\xea\xac\xba\x2b\x2c\xe3\x89\x3d\x73\xfe\x55\xb7\xda\x64\xbf\xa8\xa8\x6c\xec\xf9\x12\xe2\xc8\x17\x8d\x10\x41\x00\xf7\xc8\xcb\xa7\x6e\x9f\xf7\x19\x94\x98\xa3\xe1\x4b\xe4\x09\x15\xc7\x35\xf6\x8b\xf5\x46\xd8\xb2\x2d\x68\x1c\x66\xe8\x1e\x69\xef\x3f\x57\xd6\xec\x35\xe9\x63\xd9\x7f\xed\x7a\xf6\x56\x9c\x90\xe7\xb6\xfd\x71\xdf\x17\xdf\xdd\x32\x62\x15\x26\x3c\x3c\x2c\x59\x3a\xc7\x26\xfe\xfe\x8f\x80\x07\x62\x3e\x41\x4f\xdd\xfd\xd7\x02\xaf\x85\x0f\x50\xd7\xef\x21\x4d\x86\x92\xdf\x75\x94\xa3\xfa\x58\x71\x71\x8f\x76\x9f\x41\xd3\x38\x55\x7b\x90\x08\x3e\xdc\x40\x1c\x29\x57\xe8\x5b\xf1\x67\xf4\xa0\x84\x41\xd7\x9e\xc7\xc1\x9b\x2f\xae\xcf\xf0\xf0\x30\x72\xe4\x51\x4e\xc4\x33\xed\xd2\x40\x15\x1a\x6d\xbd\x79\x49\xa3\xed\xc4\xe5\xbe\x1b\xcc\x4b\x1c\x73\xff\x94\x23\xf4\x4f\x27\x34\xaf\x85\xfd\x8f\xc8\xc6\x2d\xd0\x2e\x4d\x1c\xc9\x9b\x1b\xb0\x69\xde\xf9\xdf\x9e\x6e\x14\x5d\x2c\xa5\xfa\x91\xf2\x53\xc8\xda\x64\x1e\x12\xf9\x62\xd7\x88\x5d\x27\xb0\x69\xbe\xb5\x7d\xfd\x78\x5d\x79\x09\x5f\x38\x5a\x57\x5d\x50\x80\xdf\x9a\x80\x56\xf8\xc5\xf6\x38\x8a\x65\xa4\x93\x2d\x5d\x8f\xf5\xa5\x1b\x3d\x1b\xc8\x91\x7a\x3e\x94\x5b\x99\xdb\x34\x97\xdb\xc1\xbf\x25\xd9\x85\xff\x37\x00\x00\xff\xff\xaf\xd2\x3e\x70\x22\x06\x00\x00")

func templateMethodTmplBytes() ([]byte, error) {
	return bindataRead(
		_templateMethodTmpl,
		"template/method.tmpl",
	)
}

func templateMethodTmpl() (*asset, error) {
	bytes, err := templateMethodTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "template/method.tmpl", size: 1570, mode: os.FileMode(0644), modTime: time.Unix(1709174476, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xf4, 0x48, 0x2c, 0x8f, 0xe7, 0x39, 0x26, 0x5a, 0x5a, 0x94, 0xeb, 0x26, 0x38, 0xcc, 0xe4, 0x3e, 0x87, 0x7a, 0x8e, 0x78, 0xe6, 0xde, 0x4, 0xba, 0xf9, 0xc0, 0x62, 0x93, 0x2c, 0xa7, 0xab, 0xa0}}
	return a, nil
}

var _templateStructTmpl = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x8e\x41\x0a\xc2\x30\x10\x45\xd7\xe6\x14\xff\x02\xa6\x87\x10\x5c\xba\x69\x2f\x10\xdb\x8f\x0a\x49\x2c\x49\xba\x90\x61\xee\x2e\x23\xb5\x08\x76\xf7\x87\x3f\xef\xcd\x74\x1d\x44\xe0\xfb\x56\x96\xb1\xf9\x4b\x48\x84\x2a\x0a\xe7\xc2\xca\xdc\xea\x6f\x3b\x84\x6b\xa4\xef\xc7\x3b\x53\x80\xaa\xff\xab\x56\xdc\xb5\xd7\xcc\x3d\x6d\xfd\x8c\x10\x27\x72\x44\x09\xf9\xc6\x6d\xe5\xfc\x60\x9c\xaa\xc1\x07\x03\xbf\x84\xe5\xc1\x6c\xaa\x58\x5f\x3d\x3d\xe3\x92\xf2\x76\xcb\x54\xcc\x93\x45\x75\xef\x00\x00\x00\xff\xff\xc1\x7f\x9b\xb1\xce\x00\x00\x00")

func templateStructTmplBytes() ([]byte, error) {
	return bindataRead(
		_templateStructTmpl,
		"template/struct.tmpl",
	)
}

func templateStructTmpl() (*asset, error) {
	bytes, err := templateStructTmplBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "template/struct.tmpl", size: 206, mode: os.FileMode(0644), modTime: time.Unix(1709174476, 0)}
	a := &asset{bytes: bytes, info: info, digest: [32]uint8{0xf2, 0xb0, 0xa8, 0xc6, 0xb6, 0xc9, 0xaa, 0x8, 0xbf, 0x87, 0x8, 0x14, 0x6b, 0x4a, 0xfb, 0x62, 0xbb, 0xc, 0x9f, 0x84, 0x8, 0xae, 0x80, 0xdb, 0x5f, 0xce, 0xf6, 0x8d, 0xee, 0x25, 0x90, 0x32}}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// AssetString returns the asset contents as a string (instead of a []byte).
func AssetString(name string) (string, error) {
	data, err := Asset(name)
	return string(data), err
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// MustAssetString is like AssetString but panics when Asset would return an
// error. It simplifies safe initialization of global variables.
func MustAssetString(name string) string {
	return string(MustAsset(name))
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetDigest returns the digest of the file with the given name. It returns an
// error if the asset could not be found or the digest could not be loaded.
func AssetDigest(name string) ([sha256.Size]byte, error) {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[canonicalName]; ok {
		a, err := f()
		if err != nil {
			return [sha256.Size]byte{}, fmt.Errorf("AssetDigest %s can't read by error: %v", name, err)
		}
		return a.digest, nil
	}
	return [sha256.Size]byte{}, fmt.Errorf("AssetDigest %s not found", name)
}

// Digests returns a map of all known files and their checksums.
func Digests() (map[string][sha256.Size]byte, error) {
	mp := make(map[string][sha256.Size]byte, len(_bindata))
	for name := range _bindata {
		a, err := _bindata[name]()
		if err != nil {
			return nil, err
		}
		mp[name] = a.digest
	}
	return mp, nil
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"template/method.tmpl": templateMethodTmpl,
	"template/struct.tmpl": templateStructTmpl,
}

// AssetDebug is true if the assets were built with the debug flag enabled.
const AssetDebug = false

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//
//	data/
//	  foo.txt
//	  img/
//	    a.png
//	    b.png
//
// then AssetDir("data") would return []string{"foo.txt", "img"},
// AssetDir("data/img") would return []string{"a.png", "b.png"},
// AssetDir("foo.txt") and AssetDir("notexist") would return an error, and
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		canonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(canonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"template": {nil, map[string]*bintree{
		"method.tmpl": {templateMethodTmpl, map[string]*bintree{}},
		"struct.tmpl": {templateStructTmpl, map[string]*bintree{}},
	}},
}}

// RestoreAsset restores an asset under the given directory.
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = os.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	return os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
}

// RestoreAssets restores an asset under the given directory recursively.
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	canonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(canonicalName, "/")...)...)
}
