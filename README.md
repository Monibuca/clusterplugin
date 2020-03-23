# clusterplugin
cluster plugin for monibuca
实现了基本的集群功能，里面包含一对发布者和订阅者，分别在主从服务器中启用，进行连接。 起基本原理就是，在主服务器启动端口监听，从服务器收到播放请求时，如果从服务器没有对应的发布者，则向主服务器发起请求，主服务器收到来自从服务器的请求时，将该请求作为一个订阅者。从服务器则把tcp连接作为发布者，实现视频流的传递过程。

# 插件名称

Cluster

# 配置

源服务器的配置是ListenAddr，用来监听从服务器的请求。 边缘服务器的配置是OriginServer,表示源服务器的地址。 当然服务器可以既是源服务器也是边缘服务器，即充当中转站。

## 边缘服务器的配置
```toml
[Cluster]
OriginServer = "localhost:2019"

```

## 源服务器的配置
```toml
[Cluster]
ListenAddr = ":2019"
```

# 示意图
![sketch](https://raw.githubusercontent.com/Monibuca/clusterplugin/master/sketch.svg)