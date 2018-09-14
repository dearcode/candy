package store

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/juju/errors"
	"github.com/syndtr/goleveldb/leveldb"

	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

const (
	maxBlockSize  = 67108864
	closeInterval = time.Second * 5
)

type fileInfo struct {
	Key    []byte
	Name   string
	Offset int32
	Size   int32
	Time   int64
}

type cacheFile struct {
	name string
	last time.Time
	file *os.File
}

type fileDB struct {
	root string

	cache  map[string]*cacheFile
	db     *leveldb.DB // 所有消息都存在这里
	master *cacheFile
	sync.RWMutex
}

func newFileDB(dir string) *fileDB {
	return &fileDB{root: dir, cache: make(map[string]*cacheFile)}
}

func (f *fileDB) newFile() error {
	err := syscall.EEXIST

	for ; err == syscall.EEXIST; time.Sleep(time.Millisecond) {
		name := fmt.Sprintf("%s/%s/%x.dat", f.root, util.FileBlockPath, time.Now().UnixNano())
		file, err := os.Create(name)
		if err == nil {
			f.master = &cacheFile{name: name, file: file, last: time.Now().Add(closeInterval)}
			f.cache[name] = f.master
			return nil
		}
		log.Debugf("create file %s error:%s", name, err)
	}

	return err
}

func (f *fileDB) close(name string) {
	if c, ok := f.cache[name]; ok {
		c.last = time.Now().Add(closeInterval)
	}
}

func (f *fileDB) open(name string) (file *os.File, err error) {
	c, ok := f.cache[name]
	if ok {
		c.last = time.Now().Add(closeInterval)
		return c.file, nil
	}

	if file, err = os.Open(name); err == nil {
		f.cache[name] = &cacheFile{name: name, file: file, last: time.Now().Add(closeInterval)}
	}
	return
}

// closeFile 自动关闭长时间不操作的文件.
func (f *fileDB) closeFile() {
	for ; ; time.Sleep(time.Second) {
		var names []string
		now := time.Now()
		f.RLock()
		for n, c := range f.cache {
			if c.name != f.master.name && now.After(c.last) {
				names = append(names, n)
			}
		}
		f.RUnlock()

		for _, n := range names {
			f.Lock()
			c := f.cache[n]
			delete(f.cache, n)
			f.Unlock()

			c.file.Close()
			log.Debugf("close file:%s", c.file.Name())
		}
	}
}

func (f *fileDB) start() error {
	var err error

	path := fmt.Sprintf("%s/%s", f.root, util.FileDBPath)
	log.Debugf("path:%v", path)
	if f.db, err = leveldb.OpenFile(path, nil); err != nil {
		return errors.Trace(err)
	}

	if err = os.MkdirAll(fmt.Sprintf("%s/%s", f.root, util.FileBlockPath), os.ModePerm); err != nil {
		return errors.Trace(err)
	}

	go f.closeFile()

	return nil
}

// check master file size
func (f *fileDB) isFull() bool {
	if f.master == nil {
		return true
	}

	i, err := f.master.file.Stat()
	if err != nil {
		f.close(f.master.name)
		return true
	}

	if i.Size() > maxBlockSize {
		f.close(f.master.name)
		return true
	}

	return false
}

func (f *fileDB) add(key []byte, data []byte) error {
	f.Lock()
	defer f.Unlock()

	if f.isFull() {
		if err := f.newFile(); err != nil {
			return errors.Trace(err)
		}
	}

	offset, err := f.master.file.Seek(0, io.SeekEnd)
	if err != nil {
		return errors.Trace(err)
	}

	size, err := f.master.file.Write(data)
	if err != nil {
		return errors.Trace(err)
	}

	i := fileInfo{Key: key, Name: f.master.name, Time: time.Now().Unix(), Offset: int32(offset), Size: int32(size)}
	if data, err = json.Marshal(i); err != nil {
		return errors.Trace(err)
	}

	if err = f.db.Put(i.Key, data, nil); err != nil {
		return errors.Trace(err)
	}

	log.Debugf("add key:%x to file:%s offset:%d size:%d", key, i.Name, i.Offset, i.Size)
	return nil
}

func (f *fileDB) get(key []byte) (data []byte, err error) {
	var file *os.File
	var i fileInfo

	if data, err = f.db.Get(key, nil); err != nil {
		return
	}

	if err = json.Unmarshal(data, &i); err != nil {
		return
	}

	f.Lock()
	defer f.Unlock()

	if file, err = f.open(i.Name); err != nil {
		return
	}

	data = make([]byte, i.Size)

	_, err = file.ReadAt(data, int64(i.Offset))
	return
}

func (f *fileDB) exist(key []byte) (bool, error) {
	return f.db.Has(key, nil)
}
