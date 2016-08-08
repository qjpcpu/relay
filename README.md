relay
=====================================

简单ssh登录管理器(实际上可作为命令映射器),安装配置好后直接运行`relay`即可

### 安装

```
go get -u github.com/qjpcpu/relay
```

### 配置

配置文件`~/.relay.conf`是yaml格式:

```
- 
 name: 测试服务器
 cmd: ssh root@test.example.com
- 
 name: 线上服务器1
 cmd: ssh work@online1.example.com
-
 name: 线上服务器2
 cmd: ssh work@online2.example.com
```

### 运行

```
relay
```

![snapshot](https://raw.githubusercontent.com/qjpcpu/relay/master/snapshot.png)
