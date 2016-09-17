package store

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/juju/errors"
	"github.com/syndtr/goleveldb/leveldb"
	lu "github.com/syndtr/goleveldb/leveldb/util"

	"github.com/dearcode/candy/server/meta"
	"github.com/dearcode/candy/server/util"
	"github.com/dearcode/candy/server/util/log"
)

var (
	pushTimeout = time.Second * 10
)

type sender interface {
	send(meta.Message) error
}

// 消息处理流程
// 1.先插入到stable中，然后插入到queue中，之后返回给gate.
// 2.异步启动推送，在消息推送到所有用户之后，把消息ID从queue中删除
// 3.每次启动要检测queue中是否有未推送的消息.
// 4.定时检测是否有之前推送失败的消息，重新推送.
type messageDB struct {
	root   string
	queue  *leveldb.DB // 存储新到的消息
	stable *leveldb.DB // 所有消息都存在这里
	sender sender
	retry  map[int64]time.Time // 对应queue的内存映射
	sync.Mutex
}

func newMessageDB(dir string) *messageDB {
	return &messageDB{root: dir, retry: make(map[int64]time.Time)}
}

func (m *messageDB) start(s sender) error {
	var err error
	if m.stable, err = leveldb.OpenFile(fmt.Sprintf("%s/%s", m.root, util.MessageDBPath), nil); err != nil {
		return errors.Trace(err)
	}
	if m.queue, err = leveldb.OpenFile(fmt.Sprintf("%s/%s", m.root, util.MessageLogDBPath), nil); err != nil {
		return errors.Trace(err)
	}

	start := util.EncodeInt64(0)
	end := util.EncodeInt64(math.MaxInt64)

	it := m.queue.NewIterator(&lu.Range{Start: start, Limit: end}, nil)
	for it.Next(); it.Valid(); it.Next() {
		m.addRetry(util.DecodeInt64(it.Key()))
	}

	m.sender = s

	go m.repush()

	return nil
}

// 定时推送之前失败过的消息
func (m *messageDB) repush() {
	for ; ; time.Sleep(time.Second) {
		var ids []int64
		now := time.Now()

		m.Lock()
		for id, tm := range m.retry {
			if now.After(tm) {
				ids = append(ids, id)
			}
		}
		m.Unlock()

		for _, id := range ids {
			msgs, err := m.get(id)
			if err != nil {
				log.Errorf("get message:%d error:%s", id, errors.ErrorStack(err))
				m.addRetry(id)
				continue
			}
			if err = m.sender.send(msgs[0]); err != nil {
				log.Errorf("push message:%+v error:%s", msgs[0], errors.ErrorStack(err))
				m.addRetry(id)
				continue
			}
			delete(m.retry, id)
			log.Debugf("remove retry message:%d", id)
		}
	}
}

func (m *messageDB) addRetry(id int64) {
	m.Lock()
	m.retry[id] = time.Now().Add(pushTimeout)
	m.Unlock()
}

func (m *messageDB) delRetry(id int64) {
	m.Lock()
	delete(m.retry, id)
	m.Unlock()
}

// 向数据库中插入新的消息.
func (m *messageDB) add(msg meta.Message) error {
	buf, err := json.Marshal(msg)
	if err != nil {
		return errors.Trace(err)
	}

	return m.stable.Put(util.EncodeInt64(msg.ID), buf, nil)
}

// 向未发送队列中添加记录
func (m *messageDB) addQueue(id int64) error {
	key := util.EncodeInt64(id)
	return m.queue.Put(key, []byte(""), nil)
}

func (m *messageDB) delQueue(id int64) {
	key := util.EncodeInt64(id)
	m.queue.Delete(key, nil)
}

func (m *messageDB) get(ids ...int64) ([]meta.Message, error) {
	var mss []meta.Message
	var msg meta.Message

	for _, id := range ids {
		v, err := m.stable.Get(util.EncodeInt64(id), nil)
		if err != nil {
			return nil, errors.Trace(err)
		}

		if err = json.Unmarshal(v, &msg); err != nil {
			return nil, errors.Trace(err)
		}

		mss = append(mss, msg)
	}

	return mss, nil
}
