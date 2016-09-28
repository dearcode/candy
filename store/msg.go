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

	"github.com/dearcode/candy/meta"
	"github.com/dearcode/candy/util"
	"github.com/dearcode/candy/util/log"
)

var (
	pushTimeout = time.Second * 10
	repushTime  = time.Millisecond * 10
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
	path := fmt.Sprintf("%s/%s", m.root, util.MessageDBPath)
	if m.stable, err = leveldb.OpenFile(path, nil); err != nil {
		log.Errorf("db openFile:%v, err:%s", path, err.Error())
		return errors.Trace(err)
	}
	path = fmt.Sprintf("%s/%s", m.root, util.MessageLogDBPath)
	if m.queue, err = leveldb.OpenFile(path, nil); err != nil {
		log.Errorf("db openFile:%v, err:%s", path, err.Error())
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
	for ; ; time.Sleep(repushTime) {
		var ids []int64
		now := time.Now()

		m.Lock()
		for id, tm := range m.retry {
			if now.Before(tm) {
				ids = append(ids, id)
			}
		}
		m.Unlock()

		for _, id := range ids {
			t := time.Now().Unix()
			log.Debugf("debug 1, time:%v", t)

			msgs, err := m.get(id)
			if err != nil {
				log.Errorf("get message:%d error:%s", id, errors.ErrorStack(err))
				m.addRetry(id)
				continue
			}

			log.Debugf("sender send msg:%v", msgs[0])
			if err = m.sender.send(*msgs[0]); err != nil {
				if errors.Cause(err) != ErrInvalidSender {
					log.Errorf("push message:%+v error:%s", msgs[0], errors.ErrorStack(err))
					//TODO 这块应该考虑使用LRU算法，把不在线的放在最后去重试，优先推送在线的消息，离线用户多的话可以增加
					//消息推送的效率???
					m.addRetry(id)
					continue
				}
			}
			//推送成功删除消息
			m.delRetry(id)
			log.Debugf("debug 2, time:%v", time.Now().Unix()-t)
			log.Debugf("remove retry message:%d", id)
		}
	}
}

func (m *messageDB) addRetry(id int64) {
	log.Debugf("id:%v", id)
	m.Lock()
	m.retry[id] = time.Now().Add(pushTimeout)
	m.Unlock()
}

func (m *messageDB) delRetry(id int64) {
	log.Debugf("id:%v", id)
	m.Lock()
	delete(m.retry, id)
	m.Unlock()
}

// 向数据库中插入新的消息.
func (m *messageDB) add(msg meta.Message) error {
	buf, err := json.Marshal(msg)
	log.Debugf("buf:%v", string(buf))
	if err != nil {
		return errors.Trace(err)
	}

	return m.stable.Put(util.EncodeInt64(msg.ID), buf, nil)
}

// 向未发送队列中添加记录
func (m *messageDB) addQueue(id int64) error {
	log.Debugf("id:%v", id)
	key := util.EncodeInt64(id)

	m.addRetry(id)
	return m.queue.Put(key, []byte(""), nil)
}

func (m *messageDB) send(msg meta.Message) error {
	//直接发送，如果发送失败再插入队列中
	if err := m.sender.send(msg); err != nil {
		if err := m.addQueue(msg.ID); err != nil {
			return errors.Trace(err)
		}
	}
	return nil
}

func (m *messageDB) delQueue(id int64) {
	log.Debugf("id:%v", id)
	key := util.EncodeInt64(id)
	m.queue.Delete(key, nil)
}

func (m *messageDB) get(ids ...int64) ([]*meta.Message, error) {
	log.Debugf("ids:%v", ids)
	var mss []*meta.Message

	for _, id := range ids {
		v, err := m.stable.Get(util.EncodeInt64(id), nil)
		if err != nil {
			return nil, errors.Trace(err)
		}

		var msg meta.Message
		if err = json.Unmarshal(v, &msg); err != nil {
			return nil, errors.Trace(err)
		}

		mss = append(mss, &msg)
	}

	return mss, nil
}
