# Esim文档

## 架构
![此处输入图片的描述][1]

## 分层

![此处输入图片的描述][2]

### 各层职责

目录 | 职责
---|---
controller | 负责显示信息和解析、校验请求，适配不同终端
app | 不包含业务规则，为下一层领域模型协调任务，分配工作
domain| 负责表达业务概念，业务状态信息和业务规则，是业务软件的核心
infrastructure|为各层提供技术支持，持久化，领域事件等

### 编码规范

- [Uber Style 编码规范](https://github.com/xxjwxc/uber_go_guide_cn)

- 函数第一个参数是ctx！函数第一个参数是ctx！函数第一个参数是ctx！

- 命名
&emsp; | Jaeger
---|---
目录名 |小写/中横线
函数名 |小驼峰
文件名 |下划线
变量| 小驼峰
常量|小驼峰
包名|当前目录名
请求地址|*小写
请求参数|小驼峰
返回参数|小驼峰

目录 |定义|文件| 类 | 接口
---|---|---|---
app |应用层|coupon.go | CouponService|无
domain/service|领域服务 | coupon.go | CouponService|无
domain/entity |实体| coupon.go | Coupon|无
infra/event |领域事件|coupon.go | couponEvent | CouponEvent
infra/repo|资源库|coupon.go| couponRepo |CouponRepo
infra/dao|数据访问对象| coupon.go| CouponDao |无

### 数据库设计规范小三样

```mysql
`create_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
`last_update_time` TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
`is_deleted` TINYINT(1) UNSIGNED NOT NULL DEFAULT '0' COMMENT '删除标识',
```


## 特性

- 由三层架构演进为四层架构（DDD + 简洁架构）
- 面向接口编程
- 编译时的依赖注入
- 管控业务使用的网络io
- 融入log，opentracing，metrice提升服务可观察性
- 单元测试友好，面向TDD


## 依赖注入
> Esim 使用[wire](https://github.com/google/wire)实现编译时的依赖注入，它有以下优点：

- 当依赖关系图变得复杂时，运行时依赖注入很难跟踪和调试。 使用代码生成意味着在运行时执行的初始化代码是常规的，惯用的Go代码，易于理解和调试。不会因为框架的各种奇技淫巧而变得生涩难懂。特别重要的是，忘记依赖项等问题会成为编译时错误，而不是运行时错误。
- 与服务定位器不同，不需要费心编造名称来注册服务。 Wire使用Go语法中的类型将组件与其依赖项连接起来。
- 更容易防止依赖项变得臃肿。Wire生成的代码只会导入您需要的依赖项，因此您的二进制文件将不会有未使用的导入。 运行时依赖注入在运行之前无法识别未使用的依赖项。
- Wire的依赖图是静态可知的，这为工具化和可视化提供了可能。

> Esim将wire用于业务与基础设施之间。将基础设施的初始化从业务抽离出来，集中管理。

### Esim使用wire示例
> 基础设置的依赖和初始化都在 ```infra/infra.go``` 文件下。wire的使用主要分2步，以增加mysqlClient：

#### provide
##### before

```golang
type Infra struct {
	*container.Esim
}

var infraSet = wire.NewSet(
	wire.Struct(new(Infra), "*"),
	provideEsim,
)
```

##### after

```golang

type Infra struct {
	*container.Esim

	DB mysql.MysqlClient
}

var infraSet = wire.NewSet(
	wire.Struct(new(Infra), "*"),
	provideEsim,
	provideDb,
)

func provideDb(esim *container.Esim) mysql.MysqlClient {
    ......
	return mysqlClent
}
```

#### Inject

> 在当前目录下执行：```wire```命令，看到：

```linux
wire: projectPath/internal/infra: wrote projectPath/internal/infra/wire_gen.go
```
> 说明执行成功，就可以在项目中使用了。

```golang
infra.NewInfra().DB
```

## 依赖倒置
> 依赖倒置和依赖注入一样，都是应用于业务与基础设施之间。主要的作用是让业务与技术实现分离。
> 在实际的使用中我们把涉及io操作都放到了基础设施的资源库上。这样做的好处：

- 单元测试变简单，使用mock代替数据源
- 不用学习各种 mock sdk，只针对 app 和 domain写单元测试
- 不依赖远程，可以单机进行开发

## 工具

- esim db2entity -d db_name -d table_name

> 前置条件：
> 1. 在项目根目录下

> 由于DDD开发方式多了很多目录，文件，导致这部分工作变得很繁琐，所以```db2entity``` 从数据库的表开始，自动建立实体，生成简单的CRUD语句和资源库的接口与实现，并把生成的资源库注入到基础设施。

- esim model -m modelname

> 前置条件:
> 1. 在模型目录下
> 2. 建议关闭module， ```export GO111MODULE=off```
> 3. 先备份模型所在文件

> 注意：
> 1. 受限于go的[plugin](https://mojotv.cn/go/golang-plugin-tutorial)机制，只支持linux/mac
> 2. 优化的模型不能有太多第三方依赖，正常情况下可以满足业务的model优化。

> 当项目进入到调优阶段，由于DDD将模型和数据分离，可以单独对模型进行优化。```model``` 命令可以自动对模型进行初始化，内存对齐，生成临时对象池，reset和释放model。很大程度的减少调优花费的时间和心智负担。

## 安装

> 环境 go 1.2 及以上

> 使用 module 包管理工具

> go get github.com/jukylin/esim

## 创建项目

```golang
esim new -s test
cd test
```


## 启动Gin服务
```golang
go run main.go gin
```

### 访问

> http://localhost:8080



## 启动GRPC服务
```
go run main.go grpc
```


## 配置
- 配置文件

> 配置文件在项目的conf目录下，分3个环境配置文件

> conf/dev.yaml

> conf/test.yaml

> conf/pro.yaml

- provide

```golang
func provideConf(){
    options := config.ViperConfOptions{}
    env := os.Getenv("ENV")
    if env == "" {
    	env = "dev"
    }

    file := []string{"conf/monitoring.yaml", "conf/" + env + ".yaml"}
    conf := config.NewViperConfig(options.WithConfigType("yaml"),
    	options.WithConfFile(file))

    return conf
}
```

- reference

```golang
service_name := infra.NewInfra().Conf.GetString("appname")
```

## 日志
> 日志会根据不同环境打印，开发和测试环境会把所以日志打印到终端，生产只会打印warn及以上的日志。
> Esim提供了2套日志接口，一套没有上下文，一套有。使用上下文是为了把分布式环境下的日志通过tracer_id串起来。

- provide

```golang
func provideLogger(conf config.Config) log.Logger {
	var loggerOptions log.LoggerOptions

	logger := log.NewLogger(
		loggerOptions.WithConf(conf),
		loggerOptions.WithDebug(conf.GetBool("debug")),
	)
	return logger}
```

- reference

```golang
infra.NewInfra().Log.Infof("info %s", "test")

infra.NewInfra().Log.Infoc(ctx, "info %s", "test")
```



## HTTP
> 比官方接口多了上下文（ctx）参数

- provide

```golang
func provideHttp(esim *container.Esim) http.HttpClient {
	proxyTransportOptions := http.ProxyTransportOptions{}

	proxyTransport := http.NewProxyTransport(
		proxyTransportOptions.WithConf(esim.Conf),
		proxyTransportOptions.WithLogger(esim.Log),
	)

	clientOptions := http.ClientOptions{}
	httpClent := http.NewHttpClient(
		clientOptions.WithTransport(proxyTransport),
		clientOptions.WithTimeOut(esim.Conf.GetDuration("http_client_time_out")),
	)

	return httpClent
}
```

- reference

```golang
import "gitlab.etcchebao.cn/go_service/esim/pkg/http"

resp, err := nfra.NewInfra().Http.GetCtx(ctx, "http://www.baidu.com")
```


## Mongodb
> 文档 https://github.com/mongodb/mongo-go-driver

- provide

```golang
func provideMongodb(esim *container.Esim) mongodb.MgoClient {
	eventOptions := mongodb.EventOptions{}
	options := mongodb.MgoClientOptions{}
	mongo := mongodb.NewMongo(
		options.WithConf(esim.Conf),
		options.WithLogger(esim.Log),
		options.WithMongoEvent(mongodb.NewMonitor),
		options.WithEventOptions(eventOptions.WithMonConf(esim.Conf),
			eventOptions.WithMonLogger(esim.Log)),
	)

	return mongo
}
```

- reference

```golang

import "go.mongodb.org/mongo-driver/bson"

type Info struct{
	Title string
}

inf := infra.NewInfra()

info := Info{}

coll := inf.Mgo.GetColl("database", "coll")
filter := bson.M{"phone": "123456"}
res := coll.FindOne(inf.Mgo.GetCtx(c.Request.Context()), filter).Decode(&info)

```

## GRPC
> 文档 https://github.com/grpc/grpc-go

- provide

```golang
func provideGrpcClient(esim *container.Esim) grpc.GrpcClient {

	options := grpc.ClientOptions{}
	grpcClient := grpc.NewGrpcClient(
		options.WithClientConf(esim.Conf),
		options.WithClientLogger(esim.Log))

	return grpcClient
}
```

- reference

```golang
import (
    "pathto/protobuf/passport"
)

conn := infra.NewInfra().Grpc.DialContext(ctx, ":60080")
defer conn.Close()

client := passport.NewUserInfoClient(conn)

getUserByUserNameRequest := &passport.GetUserByUserNameRequest{}
getUserByUserNameRequest.Username = "123456"

replyData, err = client.GetUserByUserName(ctx, getUserByUserNameRequest)
```

## Redis
> 文档 https://github.com/gomodule/redigo

- provide

```golang
func provideRedis(esim *container.Esim) redis.RedisClient {
	proxyConnOptions := redis.ProxyConnOptions{}

	proxyConn := redis.NewProxyConn(
		proxyConnOptions.WithConf(esim.Conf),
		proxyConnOptions.WithLogger(esim.Log))

	redisClientOptions := redis.RedisClientOptions{}
	redisClent := redis.NewRedisClient(
		redisClientOptions.WithConf(esim.Conf),
		redisClientOptions.WithLogger(esim.Log),
		redisClientOptions.WithProxyConn(proxyConn),
	)

	return redisClent
}
```

- reference

```golang

"gitlab.etcchebao.cn/go_service/esim/pkg/redis"

conn := infra.NewInfra().Redsi.GetCtxRedisConn(ctx)
defer conn.Close()
key := "username:"+username
exists, err := redis.Bool(conn.Do("exists", key))
```

## Mysql
> 文档 https://gorm.io/docs/

- provide

```golang
func provideDb(esim *container.Esim) mysql.MysqlClient {
	proxySqlCommonOptions := mysql.ProxySqlCommonOptions{}

	proxy := mysql.NewProxySqlCommon
	mysqlClientOptions := mysql.MysqlClientOptions{}
	mysqlClent := mysql.NewMysqlClient(
		mysqlClientOptions.WithConf(esim.Conf),
		mysqlClientOptions.WithLogger(esim.Log),
		mysqlClientOptions.WithProxySqlCommon(proxy),
		mysqlClientOptions.WithProxySqlCommonOptions(
			proxySqlCommonOptions.WithLogger(esim.Log),
			proxySqlCommonOptions.WithConf(esim.Conf)),
	)

	return mysqlClent
}
```

- reference

```golang
var user model.User
infra.NewInfra().DB.GetDb(ctx, "db").Table("table").Where("username = ?", username).
	Select([]string{"id"}).First(&user)
```




  [1]: https://imgconvert.csdnimg.cn/aHR0cHM6Ly9hdGEyLWltZy5jbi1oYW5nemhvdS5vc3MtcHViLmFsaXl1bi1pbmMuY29tL2EzM2I4MGJjYWM1ZWM3M2QwZDEzNThkNmI0OWExMTljLnBuZw?x-oss-process=image/format,png
  [2]: https://upload.cc/i1/2019/12/15/ZVh3iL.png