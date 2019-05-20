package database

import (
	"strconv"
	"time"

	"github.com/amsalt/log"
	"github.com/amsalt/netkit/util"
	"github.com/amsalt/nginet/core"
	"github.com/go-redis/redis"
)

var Redis *RedisClient

const (
	DB_RESULT_SUCCESS = iota
	DB_RESULT_FAIL
	DB_RESULT_MISS
)

const (
	REDIS_GET = iota
	REDIS_SET
	REDIS_SETNX
	REDIS_DEL
	REDIS_KEYS
	REDIS_HSET
	REDIS_HGET
	REDIS_HMSET
	REDIS_HMGET
	REDIS_HGETALL
	REDIS_HDEL
	REDIS_HINCRBY
	REDIS_LPUSH
	REDIS_LPOP
	REDIS_RPUSH
	REDIS_RPOP
	REDIS_LRANGE
	REDIS_ZADD
	REDIS_ZREVRANK
	REDIS_ZREVRANGE
	REDIS_ZINCRBY
	REDIS_SADD
	REDIS_SMEMBERS
	REDIS_SISMEMBER
	REDIS_SRANDMEMBER
	REDIS_SREM
	REDIS_PIPELINED
	REDIS_EXPIRE
	REDIS_EXPIRE_AT
	REDIS_ZREM
)

type Callback func(key string, val string, result int)
type ArrRetCallback func(key string, val []interface{}, result int)
type StrArrRetCallback func(key string, val []string, result int)
type MapRetCallback func(key string, val map[string]string, result int)
type ZArrRetCallback func(key string, z []redis.Z, result int)
type BoolCallback func(key string, val bool, result int)
type PipelineFunctor func(pipeline redis.Pipeliner) error
type PipelineCallback func(key string, cmders []redis.Cmder, result int)

type RedisCommand struct {
	action       int
	key          string
	mfields      []string
	val          string
	boolVal      bool
	expiration   time.Duration
	expirationAt time.Time
	obj          interface{}
	mvals        map[string]interface{}
	lvals        []interface{}
	svals        []string
	smvals       map[string]string
	start        int64
	stop         int64
	member       string
	zval         redis.Z
	zvals        []redis.Z
	pFunc        PipelineFunctor
	cmders       []redis.Cmder
	increment    float64

	cb       Callback
	arrCb    ArrRetCallback
	strArrCb StrArrRetCallback
	mapCb    MapRetCallback
	zArrCb   ZArrRetCallback
	bCb      BoolCallback
	pCb      PipelineCallback

	result int
}

type RedisOption struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

type RedisClient struct {
	redis    *redis.Client
	config   *RedisOption
	commands chan *RedisCommand

	executor core.Executor
}

// Redis pub-sub client need different from normal read-write node
// So open the create interface for not singleton client.
func NewRedisClient(redisOpts *RedisOption, executor core.Executor) *RedisClient {
	redis := &RedisClient{config: redisOpts, executor: executor}
	redis.Init()
	return redis
}

func (redisClient *RedisClient) GetRawClient() *redis.Client {
	return redisClient.redis
}

func (redisClient *RedisClient) Init() {
	redisClient.commands = make(chan *RedisCommand, DefaultChanLen)

	redisClient.redis = redis.NewClient(&redis.Options{
		Addr:     redisClient.config.Addr,
		Password: redisClient.config.Password,
		DB:       0,
		PoolSize: redisClient.config.PoolSize,
	})

	for i := 0; i < DefaultWorkerNum; i++ {
		redisClient.loop()
	}
}

func (redisClient *RedisClient) Keys(pattern string) {
	item := &RedisCommand{action: REDIS_KEYS, key: pattern}
	redisClient.pushCommand(item)
}

/*--------------------------------- Base --------------------------------------*/
func (redisClient *RedisClient) Expire(key string, expireSecs time.Duration) {
	item := &RedisCommand{action: REDIS_EXPIRE, key: key, expiration: expireSecs}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) ExpireAt(key string, expirationAt time.Time) {
	item := &RedisCommand{action: REDIS_EXPIRE_AT, key: key, expirationAt: expirationAt}
	redisClient.pushCommand(item)
}

/*--------------------------------- String --------------------------------------*/
func (redisClient *RedisClient) Set(key string, val string, expiration time.Duration) {
	item := &RedisCommand{action: REDIS_SET, key: key, val: val, expiration: expiration}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) SetNX(key string, val string, expiration time.Duration, cb BoolCallback) {
	item := &RedisCommand{action: REDIS_SETNX, key: key, val: val, expiration: expiration, bCb: cb}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) Del(key string) {
	item := &RedisCommand{action: REDIS_DEL, key: key}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) Get(key string, cb Callback) {
	item := &RedisCommand{action: REDIS_GET, key: key, cb: cb}
	redisClient.pushCommand(item)
}

/*--------------------------------- Hash --------------------------------------*/
func (redisClient *RedisClient) HSet(key string, filed string, val string) {
	fileds := []string{}
	fileds = append(fileds, filed)

	item := &RedisCommand{action: REDIS_HSET, mfields: fileds, key: key, val: val}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) HGet(key string, field string, cb Callback) {
	fileds := []string{}
	fileds = append(fileds, field)

	item := &RedisCommand{action: REDIS_HGET, key: key, mfields: fileds, cb: cb}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) HMSet(key string, mvals map[string]interface{}) {
	item := &RedisCommand{action: REDIS_HMSET, key: key, mvals: mvals}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) HMGet(key string, fileds []string, cb ArrRetCallback) {
	item := &RedisCommand{action: REDIS_HMGET, key: key, mfields: fileds, arrCb: cb}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) HGetAll(key string, cb MapRetCallback) {
	item := &RedisCommand{action: REDIS_HGETALL, key: key, mapCb: cb}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) HDel(key string, fileds []string, cb Callback) {
	item := &RedisCommand{action: REDIS_HDEL, key: key, mfields: fileds, cb: cb}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) HIncrBy(key string, field string, cb Callback) {
	fileds := []string{}
	fileds = append(fileds, field)

	item := &RedisCommand{action: REDIS_HINCRBY, key: key, mfields: fileds, cb: cb}
	redisClient.pushCommand(item)
}

/*--------------------------------- List --------------------------------------*/
func (redisClient *RedisClient) LPush(key string, vals []interface{}) {
	item := &RedisCommand{action: REDIS_LPUSH, key: key, lvals: vals}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) RPush(key string, vals []interface{}) {
	item := &RedisCommand{action: REDIS_RPUSH, key: key, lvals: vals}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) LPop(key string, cb Callback) {
	item := &RedisCommand{action: REDIS_LPOP, key: key, cb: cb}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) RPop(key string, cb Callback) {
	item := &RedisCommand{action: REDIS_RPOP, key: key, cb: cb}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) LRange(key string, start int64, stop int64, cb StrArrRetCallback) {
	item := &RedisCommand{action: REDIS_LRANGE, key: key, start: start, stop: stop, strArrCb: cb}
	redisClient.pushCommand(item)
}

/*--------------------------------- Sorted Set --------------------------------------*/
func (redisClient *RedisClient) ZAdd(key string, vals interface{}) {
	item := &RedisCommand{action: REDIS_ZADD, key: key, obj: vals}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) ZRevRank(key string, member string, cb Callback) {
	item := &RedisCommand{action: REDIS_ZREVRANK, key: key, member: member, cb: cb}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) ZRevRange(key string, start int, stop int, cb ZArrRetCallback) {
	item := &RedisCommand{action: REDIS_ZREVRANGE, key: key, start: int64(start), stop: int64(stop), zArrCb: cb}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) ZIncrBy(key string, increment float64, member string, cb Callback) {
	item := &RedisCommand{action: REDIS_ZINCRBY, key: key, increment: increment, member: member, cb: cb}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) ZRem(key string, vals ...string) {
	item := &RedisCommand{action: REDIS_ZREM, key: key, svals: vals}
	redisClient.pushCommand(item)
}

/*--------------------------------- Set --------------------------------------*/
func (redisClient *RedisClient) SAdd(key string, vals ...string) {
	item := &RedisCommand{action: REDIS_SADD, key: key, svals: vals}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) SMembers(key string, cb StrArrRetCallback) {
	item := &RedisCommand{action: REDIS_SMEMBERS, key: key, strArrCb: cb}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) SRandMember(key string, cb Callback) {
	item := &RedisCommand{action: REDIS_SRANDMEMBER, key: key, cb: cb}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) SIsMember(key string, cb BoolCallback) {
	item := &RedisCommand{action: REDIS_SISMEMBER, key: key, bCb: cb}
	redisClient.pushCommand(item)
}

func (redisClient *RedisClient) SRem(key string, vals ...string) {
	item := &RedisCommand{action: REDIS_SREM, key: key, svals: vals}
	redisClient.pushCommand(item)
}

/*------------------------------ Pipeline ----------------------------*/
func (redisClient *RedisClient) Pipeline(pFunc PipelineFunctor, cb PipelineCallback) {
	item := &RedisCommand{action: REDIS_PIPELINED, pFunc: pFunc, pCb: cb}
	redisClient.pushCommand(item)
}

/*--------------------------------- Common --------------------------------------*/
func (redisClient *RedisClient) GetSync(key string) (string, error) {
	return redisClient.redis.Get(key).Result()
}

func (redisClient *RedisClient) pushCommand(command *RedisCommand) {
	select {
	case redisClient.commands <- command:
	default:
		redisClient.commands <- command
	}
}

// loop starts new goroutine for I/O
func (redisClient *RedisClient) loop() {
	go func() {
		// Avoid crash.
		util.TryCatch(func() {
			for {
				select {
				case command := <-redisClient.commands:
					// log.Debug("new command: %+v", command)
					redisClient.process(command)
				}
			}
		}, nil)
	}()
}

// if a Executor set, run callback in executor.
// otherwise run callback in current goroutine.
func (redisClient *RedisClient) executeCb(command *RedisCommand) {
	if redisClient.executor != nil {
		redisClient.executor.Execute(func() {
			redisClient.callCommand(command)
		})
	} else {
		redisClient.callCommand(command)
	}
}

func (redisClient *RedisClient) callCommand(command *RedisCommand) {
	if command.cb != nil {
		command.cb(command.key, command.val, command.result)
	} else if command.arrCb != nil {
		command.arrCb(command.key, command.lvals, command.result)
	} else if command.strArrCb != nil {
		command.strArrCb(command.key, command.svals, command.result)
	} else if command.mapCb != nil {
		command.mapCb(command.key, command.smvals, command.result)
	} else if command.zArrCb != nil {
		command.zArrCb(command.key, command.zvals, command.result)
	} else if command.bCb != nil {
		command.bCb(command.key, command.boolVal, command.result)
	} else if command.pCb != nil {
		command.pCb(command.key, command.cmders, command.result)
	}
}

func (redisClient *RedisClient) boolResult(command *RedisCommand, val bool, err error) {
	result := checkResult(err)
	command.result = result
	command.boolVal = val
	redisClient.executeCb(command)
}

func (redisClient *RedisClient) oneResult(command *RedisCommand, val string, err error) {
	result := checkResult(err)
	command.result = result
	command.val = val
	redisClient.executeCb(command)
}

func (redisClient *RedisClient) arrResults(command *RedisCommand, val []interface{}, err error) {
	result := checkResult(err)
	command.result = result
	command.lvals = val
	redisClient.executeCb(command)
}

func (redisClient *RedisClient) strArrResults(command *RedisCommand, val []string, err error) {
	result := checkResult(err)
	command.result = result
	command.svals = val
	redisClient.executeCb(command)
}

func (redisClient *RedisClient) zArrResults(command *RedisCommand, val []redis.Z, err error) {
	result := checkResult(err)
	command.result = result
	command.zvals = val
	redisClient.executeCb(command)
}

func (redisClient *RedisClient) mapResults(command *RedisCommand, val map[string]string, err error) {
	result := checkResult(err)
	command.result = result
	command.smvals = val
	redisClient.executeCb(command)
}

func (redisClient *RedisClient) pipelineResults(command *RedisCommand, cmders []redis.Cmder, err error) {
	result := checkResult(err)
	command.result = result
	command.cmders = cmders
	redisClient.executeCb(command)
}

func checkResult(err error) int {
	result := DB_RESULT_SUCCESS
	if err != nil {
		if err == redis.Nil {
			result = DB_RESULT_MISS
		} else {
			log.Errorf("Redis error: %+v", err)
			result = DB_RESULT_FAIL
		}
	}
	return result
}

func (redisClient *RedisClient) process(command *RedisCommand) {
	switch command.action {
	case REDIS_GET:
		val, err := redisClient.redis.Get(command.key).Result()
		redisClient.oneResult(command, val, err)

	case REDIS_SET:
		redisClient.redis.Set(command.key, command.val, command.expiration)
	case REDIS_SETNX:
		val := redisClient.redis.SetNX(command.key, command.val, command.expiration)
		redisClient.boolResult(command, val.Val(), val.Err())

	case REDIS_DEL:
		redisClient.redis.Del(command.key)
	case REDIS_KEYS:
		val, err := redisClient.redis.Keys(command.key).Result()
		redisClient.strArrResults(command, val, err)

	case REDIS_HSET:
		redisClient.redis.HSet(command.key, command.mfields[0], command.val)
	case REDIS_HGET:
		val, err := redisClient.redis.HGet(command.key, command.mfields[0]).Result()
		redisClient.oneResult(command, val, err)

	case REDIS_HMSET:
		redisClient.redis.HMSet(command.key, command.mvals).Result()
		// log.Debug("handle redis command hmset: %+v %+v", str, err)
	case REDIS_HMGET:
		val, err := redisClient.redis.HMGet(command.key, command.mfields...).Result()
		redisClient.arrResults(command, val, err)

	case REDIS_HGETALL:
		val, err := redisClient.redis.HGetAll(command.key).Result()
		redisClient.mapResults(command, val, err)

	case REDIS_HDEL:
		redisClient.redis.HDel(command.key, command.mfields...)
	case REDIS_HINCRBY:
		redisClient.redis.HIncrBy(command.key, command.mfields[0], 1) // todo: current default increase step is 1
	case REDIS_LPUSH:
		redisClient.redis.LPush(command.key, command.lvals...)
	case REDIS_LPOP:
		val, err := redisClient.redis.LPop(command.key).Result()
		redisClient.oneResult(command, val, err)

	case REDIS_RPUSH:
		redisClient.redis.RPush(command.key, command.lvals...)
	case REDIS_RPOP:
		val, err := redisClient.redis.RPop(command.key).Result()
		redisClient.oneResult(command, val, err)

	case REDIS_LRANGE:
		val, err := redisClient.redis.LRange(command.key, command.start, command.stop).Result()
		redisClient.strArrResults(command, val, err)

	case REDIS_ZADD:
		redisClient.redis.ZAdd(command.key, command.obj.(redis.Z))
		// log.Debug("handle redis command zdd: %+v", err)
	case REDIS_ZREVRANK:
		val, err := redisClient.redis.ZRevRank(command.key, command.member).Result()
		// log.Debug("handle redis command zrevrank: %+v, err:%+v", val, err)
		redisClient.oneResult(command, strconv.Itoa(int(val)), err)

	case REDIS_ZREVRANGE:
		val, err := redisClient.redis.ZRevRangeWithScores(command.key, command.start, command.stop).Result()
		// log.Debug("handle redis command zrevrange: %+v, err:%+v", val, err)
		redisClient.zArrResults(command, val, err)

	case REDIS_ZINCRBY:
		val, err := redisClient.redis.ZIncrBy(command.key, command.increment, command.member).Result()
		// log.Debug("handle redis command zincrby: %+v, err: %+v", val, err)
		if command.cb != nil {
			redisClient.oneResult(command, strconv.FormatFloat(val, 'f', 6, 64), err)
		}

	case REDIS_SADD:
		s := make([]interface{}, len(command.svals))
		for i, v := range command.svals {
			s[i] = v
		}
		redisClient.redis.SAdd(command.key, s...)
		// log.Debug("handle redis command sadd: %+v", err)
	case REDIS_SMEMBERS:
		val, err := redisClient.redis.SMembers(command.key).Result()
		redisClient.strArrResults(command, val, err)

	case REDIS_SRANDMEMBER:
		val, err := redisClient.redis.SRandMember(command.key).Result()
		redisClient.oneResult(command, val, err)

	case REDIS_SISMEMBER:
		isMember := redisClient.redis.SIsMember(command.key, command.val)
		redisClient.boolResult(command, isMember.Val(), isMember.Err())

	case REDIS_SREM:
		s := make([]interface{}, len(command.svals))
		for i, v := range command.svals {
			s[i] = v
		}
		redisClient.redis.SRem(command.key, s...)
		// log.Debug("handle redis command srem: %+v", err)

	case REDIS_PIPELINED:
		cmders, err := redisClient.redis.Pipelined(command.pFunc)
		// log.Debug("handle redis command pipeline, cmders: %+v, err: %+v", cmders, err)
		redisClient.pipelineResults(command, cmders, err)

	case REDIS_EXPIRE:
		redisClient.redis.Expire(command.key, command.expiration)

	case REDIS_EXPIRE_AT:
		redisClient.redis.ExpireAt(command.key, command.expirationAt)

	case REDIS_ZREM:
		s := make([]interface{}, len(command.svals))
		for i, v := range command.svals {
			s[i] = v
		}
		redisClient.redis.ZRem(command.key, s...)
		// log.Debug("handle redis command zrem: %+v", err)

	default:
		panic("not support!")
	}
}
