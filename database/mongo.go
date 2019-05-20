package database

import (
	"github.com/amsalt/log"
	"github.com/amsalt/nginet/core"
	"github.com/amsalt/nginet/safe"
	mgo "gopkg.in/mgo.v2"
)

var Mongo *MongoClient

const (
	MONGO_FIND = iota
	MONGO_FINDALL
	MONGO_UPSERT
	MONGO_UPDATEID
	MONGO_UPDATEALL
	MONGO_DELETE
	MONGO_INSERT
	MONGO_AGGREGATE
	MONGO_DELETE_WITH_CB
)

type MongoCondition map[string]interface{}
type MongoCallback func(error)

type MongoCommand struct {
	action    int
	col       string
	condition interface{}
	record    interface{}
	err       error
	cb        MongoCallback
}

type MongoOption struct {
	Addr     string
	Name     string
	PoolSize int
}

type MongoClient struct {
	db      *mgo.Database
	session *mgo.Session

	commands chan *MongoCommand
	executor core.Executor
}

func InitMongo(opts *MongoOption, executor core.Executor) {
	Mongo = newMongoClient(opts, executor)
}

func newMongoClient(opts *MongoOption, executor core.Executor) *MongoClient {
	mongo := &MongoClient{executor: executor}
	mongo.Init(opts)
	return mongo
}

func (mongoClient *MongoClient) Init(opts *MongoOption) {
	mongoClient.commands = make(chan *MongoCommand, DefaultChanLen)

	session, err := mgo.Dial(opts.Addr)
	if err != nil {
		log.Errorf("mongo connect error %v,%v\n", err, opts.Addr)
		panic(err)
	}

	mongoClient.session = session

	mongoClient.db = session.DB(opts.Name)
	mongoClient.session.SetPoolLimit(opts.PoolSize)
	mongoClient.session.SetMode(mgo.Eventual, true)

	for i := 0; i < DefaultWorkerNum; i++ {
		mongoClient.loop()
	}
}

func (mongoClient *MongoClient) FindSync(col string, condition interface{}, record interface{}) error {
	err := mongoClient.db.C(col).Find(condition).One(record)
	return err
}

func (mongoClient *MongoClient) Find(col string, condition interface{}, record interface{}, cb MongoCallback) error {
	command := &MongoCommand{action: MONGO_FIND, col: col, condition: condition, record: record, cb: cb}
	mongoClient.pushCommand(command)
	return nil
}

func (mongoClient *MongoClient) FindAll(col string, condition interface{}, record interface{}, cb MongoCallback) error {
	command := &MongoCommand{action: MONGO_FINDALL, col: col, condition: condition, record: record, cb: cb}
	mongoClient.pushCommand(command)
	return nil
}

func (mongoClient *MongoClient) Aggregate(col string, condition interface{}, record interface{}, cb MongoCallback) error {
	command := &MongoCommand{action: MONGO_AGGREGATE, col: col, condition: condition, record: record, cb: cb}
	mongoClient.pushCommand(command)
	return nil
}

func (mongoClient *MongoClient) Upsert(col string, condition interface{}, record interface{}, cb MongoCallback) error {
	command := &MongoCommand{action: MONGO_UPSERT, col: col, condition: condition, record: record, cb: cb}
	mongoClient.pushCommand(command)
	return nil
}

func (mongoClient *MongoClient) UpdateId(col string, condition interface{}, record interface{}, cb MongoCallback) error {
	command := &MongoCommand{action: MONGO_UPDATEID, col: col, condition: condition, record: record, cb: cb}
	mongoClient.pushCommand(command)
	return nil
}

func (mongoClient *MongoClient) UpdateAll(col string, condition interface{}, record interface{}, cb MongoCallback) error {
	command := &MongoCommand{action: MONGO_UPDATEALL, col: col, condition: condition, record: record, cb: cb}
	mongoClient.pushCommand(command)
	return nil
}

func (mongoClient *MongoClient) Insert(col string, record interface{}, cb MongoCallback) error {
	command := &MongoCommand{action: MONGO_INSERT, col: col, condition: nil, record: record, cb: cb}
	mongoClient.pushCommand(command)
	return nil
}

func (mongoClient *MongoClient) UpsertSync(col string, condition interface{}, record interface{}) error {
	_, err := mongoClient.db.C(col).Upsert(condition, record)
	return err
}

func (mongoClient *MongoClient) Remove(col string, condition interface{}) error {
	command := &MongoCommand{action: MONGO_DELETE, col: col, condition: condition}
	mongoClient.pushCommand(command)
	return nil
}

func (mongoClient *MongoClient) RemoveWithCb(col string, condition interface{}, cb MongoCallback) error {
	command := &MongoCommand{action: MONGO_DELETE_WITH_CB, col: col, condition: condition, cb: cb}
	mongoClient.pushCommand(command)
	return nil
}

func (mongoClient *MongoClient) EnsureIndexKey(col string, key ...string) {
	mongoClient.db.C(col).EnsureIndexKey(key...)
}

func (mongoClient *MongoClient) EnsureIndex(col string, index mgo.Index) {
	mongoClient.db.C(col).EnsureIndex(index)
}

func (mongoClient *MongoClient) pushCommand(command *MongoCommand) {
	select {
	case mongoClient.commands <- command:
	default:
		mongoClient.commands <- command
		log.Errorf("mongo command buffer is full")
	}
}

func (mongoClient *MongoClient) loop() {
	go func() {
		safe.TryCatch(func() {
			for {
				select {
				case command := <-mongoClient.commands:
					mongoClient.process(command)
				}
			}
		}, nil)
	}()
}

func (mongoClient *MongoClient) process(command *MongoCommand) {
	switch command.action {
	case MONGO_FIND:
		command.err = mongoClient.db.C(command.col).Find(command.condition).One(command.record)
	case MONGO_FINDALL:
		command.err = mongoClient.db.C(command.col).Find(command.condition).Limit(DefaultMongoFindMaxCount).All(command.record)
	case MONGO_AGGREGATE:
		command.err = mongoClient.db.C(command.col).Pipe(command.condition).All(command.record)
	case MONGO_UPSERT:
		_, command.err = mongoClient.db.C(command.col).Upsert(command.condition, command.record)
	case MONGO_UPDATEID:
		command.err = mongoClient.db.C(command.col).UpdateId(command.condition, command.record)
	case MONGO_UPDATEALL:
		_, command.err = mongoClient.db.C(command.col).UpdateAll(command.condition, command.record)
	case MONGO_INSERT:
		command.err = mongoClient.db.C(command.col).Insert(command.record)
	case MONGO_DELETE:
		command.err = mongoClient.db.C(command.col).Remove(command.condition)
	case MONGO_DELETE_WITH_CB:
		command.err = mongoClient.db.C(command.col).Remove(command.condition)
	default:
		panic("MongoCLient: wrong action!")
	}

	mongoClient.executeCb(command)
}

func (mongoClient *MongoClient) executeCb(command *MongoCommand) {
	if mongoClient.executor != nil {
		mongoClient.executor.Execute(func() {
			mongoClient.executeCmd(command)
		})
	} else {
		mongoClient.executeCmd(command)
	}
}

func (mongoClient *MongoClient) executeCmd(command *MongoCommand) {
	if command.err != nil && command.err.Error() != "not found" {
		log.Errorf("mongodb: err %+v", command.err)
	}

	if command.cb != nil {
		safe.TryCatch(func() {
			command.cb(command.err)
		}, func() {
			log.Errorf("error when call mongo command: %+v", command)
		})
	}
}
