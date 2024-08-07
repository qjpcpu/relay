relay
=====================================

`relay` is a command collector and trigger. Delegate all your high-frequency commands to relay, then use it conviencely.

[![IMAGE ALT TEXT](http://img.youtube.com/vi/OLfZFI7A77M/0.jpg)](https://youtu.be/OLfZFI7A77M "relay")

### Install

```
go get github.com/qjpcpu/relay
```

or build from source

```
git clone git@github.com:qjpcpu/relay.git
cd relay && go build
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

keybinding in select list:

```
move previous/next command: j/k(like vim) or C-n/C-p(like emacs)  or arrow up/down
jump to first/last command: gg/G(like vim)
jump to line: lineno+gg(like vim)
scroll page up/down: C-d/C-u(like vim)
search mode: /(like vim), C-s(like emacs)
move prev/next in search mode: C-n/C-p
confirm and run selected command: Enter
select nothing and exit: q/C-c
```

#### Bash Completion

download `autocomplete/`
For bash, add to ~/.bashrc

```
source autocomplete/bash_autocomplete
```

for zsh, add to ~/.zshrc

```
source autocomplete/zsh_autocomplete
```

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
 cmd: echo 'hello {{who}}'
 alias: hi
 options:
  who: 
   - desc: best friend
     val: Andy
```

or 

```
-
 name: command with parameters
 cmd: echo 'hello {{who}}'
 alias: hi
 defaults:
  who: DANGER
```

```
relay hi Jason
# or interactively
relay hi
```

or
##### 5.confirm command before execution

```
-
 name: command with confirm
 cmd: echo DANGER
 confirm: true
```
