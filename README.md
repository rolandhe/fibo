# fibo
纯golang版本的基于snowflake算法的 id 生成器。它提供开箱即用的能力。

# quickstart
## 编译
下载完成代码后，编译：

```
   sh build.sh
```
copy fibo程序和conf目录到您的目标目录下

## 安装zookeeper
fibo是基于zookeeper来分配当前部署实例的worker id的。

## 配置
打开conf/app.yaml 或者 app-*.yaml, 不同的配置文件代码不同的运行环境。比如 本机开发使用app.yaml, app-prod.yaml代表线上环境。fibo使用系统环境变量 FiboProfile 
来确定环境， 比如 export FiboProfile=prod, 表示线上环境，如果没有指定FiboProfile环境变量，默认是app.yaml.

### 配置具体的项

* app.port, 指定当前应用端口
* zk.servers, zookeeper多个实例的ip:port, 多个使用逗号分割
* zk.sessionTimeout, zookeeper的 session timeout， 单位是秒
* fibo.maxWorkerBits, snowflake算法中worker id所占的二进制位数，推荐是3， 最多部署8个fibo实例
* fibo.maxIdcBits, snowflake算法中idc id所占的二进制位数，推荐是3， 最多在8个idc部署
* fibo.nameSpaces， 命名空间，多个命名空间以逗号分割，不同的命名空间之间生成的id可能相同，同一命名空间内不相同。命名空间用于描述业务场景，每个业务场景有自己的业务id语义
* fibo.logLevel， 日志输出级别，0, don't output, 1 output without body, 2 output with body

## 执行 ./fibo 运行
可以每次获取一个id：
```
curl localhost:8080/fibo/one
```

也可以每次获取一批id：

```
curl localhost:8080/fibo/batch?batch=8000
```

默认是在default命名空间生成id。

如果你有自己的命名空间，请使用fibo.nameSpaces指定，比如你指定了order命名空间，那么获取改空间内的id。

```
curl localhost:8080/fibo/one/order
```

也可以每次获取一批id：

```
curl localhost:8080/fibo/batch/order?batch=8000
```

## 接口
### 单个id获取

/fibo/one[/xxx], 其中xxx代表namespace
/xxx是可选的，如果没有，代表从default命名空间获取id

### 返回值
正确返回:

```
    {
        "code":200,
        "id":440360410984218624
    }

```

错误返回：
```
    {
        "code":500,
        "errMessage":"invalid id namespace"
    }
```

### 批量获取id

/fibo/batch[/xxx]?batch=88, 其中xxx代表namespace
/xxx是可选的，如果没有，代表从default命名空间获取id
query 参数 batch 必须是[1,8196]区间的一个数， 意味着 每批最多取8196个id

#### 返回值

正确返回：
```
    {
        "code":200,
        "batchIds":[
            {
                "start":440360247079993344,
                "end":440360247079997439
            },
            {
                "start":440360247080255488,
                "end":440360247080259391
            }
        ]
    }
```

批ids可能有多段组成，每一段是[start,end]数据，连续的

错误返回：
```
    {
        "code":500,
        "errMessage":"invalid id namespace"
    }
```

# snowflake 算法变化

snowflake算法使用int64表示一个id，其中最高位是符号位，不用，其他分别存储毫秒级的unix时间戳，idc id， worker id，即连续的自增数：

* 1 bit  符号位
* 41bit，毫秒级unix时间戳， 只能表示出 2 的 41次方 -1 这么多的毫秒，大概是69年，unix元年是1971年，只能撑到2040年
* 10bit， idc id + worker id
* 12bit，1毫秒内产生的连续自增数，最多 4096个数

大多数情况下 10bit的 idc id + worker id是用不完的，所以我们推荐 idc id和worker id各使用 3 bit，因此可以节省下4位给毫秒级时间戳。上面提到
41位只能支持69年，最多支撑到2040年，再给它4位，那就可以支持1000多年了。



