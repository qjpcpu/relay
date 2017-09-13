relay
=====================================

`relay` is a command collector and trigger. Delegate all your high-frequency commands to relay, then use it conviencely.

### Install

```
go get -u github.com/qjpcpu/relay
```

or build from source

```
git clone git@github.com:qjpcpu/relay.git
cd relay && godep go build
```

or download binary file

```
# for osx
wget https://raw.githubusercontent.com/qjpcpu/relay/master/release/relay.darwin -O relay && chmod +x ./relay
# for linux
wget https://raw.githubusercontent.com/qjpcpu/relay/master/release/relay.linux -O relay && chmod +x ./relay
```

### Configuration

Default config file is `~/.relay.conf`, which is yaml format. Also you can specify config file with flag `-c`:

```
-
 name: server1
 cmd: ssh jason@10.0.2.2
-
 name: server2
 cmd: ssh work@172.1.2.3
-
 name: connect to db
 cmd: mysql -uroot -proot -h 127.0.0.1
 alias: db
-
 name: show my ip
 cmd: 'curl http://ip.cn'
 alias: ip
```

### Use relay

show help with `relay --help`:

```
NAME:
   relay - command relay station

USAGE:
   relay [global options] [command alias] [arguments...]

AUTHOR:
   JasonQu <qjpcpu@gmail.com>

COMMANDS:
     !  run last command
     @  show relay history

GLOBAL OPTIONS:
   -c value    specify config file (default: "/home/ubuntu/.relay.conf")
   --help, -h  show help
```

```
relay
```

![snapshot](https://raw.githubusercontent.com/qjpcpu/relay/master/snapshot1.png)


#### Shortcut

##### 1.run last command

```
relay !
```

##### 2.view relay history

```
relay @
```

##### 3.run command by alias

take `~/.relay.conf` as example, connect to mysql database:

```
-
 name: connect to db
 cmd: mysql -uroot -proot -h 127.0.0.1
 alias: db
```

```
relay db
```

##### 4.run alias command with parameters

```
-
 name: command with parameters
 cmd: echo 'hello {{name}}'
 alias: hi
```

```
relay hi Jason
# or interactively
relay hi
```
