# springBoot 转 kratos

### controller层转成了protobuf文件
```sh
controller + request + vo --> .proto
```

### do层转成了ent schema go文件，转成了go文件
```shell
do --> .go
```

### sql-ddl转成了ent schema go文件
```shell
sql --> .go
```
# test
```shell
./java2go.exe -h
java2go 2 go                                                            
                                                                        
Usage:                                                                  
  java2go [command]                                                     
                                                                        
Available Commands:                                                     
  completion  Generate the autocompletion script for the specified shell
  ctl         Generate the protobuf code from xxxController.java        
  do          Generate the ent schema code from xxxDO.java              
  help        Help about any command                                    
  sql         Generate the ent schema code from init-schema.sql         
                                                                        
Flags:                                                                  
  -h, --help      help for java2go                                      
  -v, --version   version for java2go

Use "java2go [command] --help" for more information about a command.
```
```sh
./java2go.exe ctl ./test/ctl ./test/ctl
```
```sh
./java2go.exe do ./test/do ./test/do
```
```sh
./java2go.exe sql ./test/sql ./test/sql
```
# 遇到的问题：
## 框架
```shell
## 前提条件当然是java的代码必须是非常的规范，包括命名规范、分层规范
## go的框架分层没有springBoot成熟，所以偷懒把controller相关很多层合并统一成protobuf做规范，具体业务实现还是要写一点代码。
## java的dao层在kratos里是data层，要自己写代码的
## go的do层要结合.pb.go以及ent/data层的具体业务，可以偷懒直接用.pb.go中的定义，一路直接到data/db
## 全部转完以后，在kratos的框架进行操作，详情看kratos官网
## kratos的serivce、biz层，通过protobuf+kratos工具自动产生，同样业务实现要自己写

```
## 语法
```shell
## go的OOP没有java这么规范，没有那么多关键词去做语法约束。一律默认public
## 单纯的数据结构中设计模式问题遇到的不多：单例问题到业务代码中处理吧
## go没法处理java的重载，遇到这种情形，要么改java函数名字，要么重名的函数自动加点东西
## java/DDL很多变量类型、注解类型，只解析了开发过程遇到的情况，没有完全覆盖所有的情况

```

## reference:
[Kratos](https://github.com/go-kratos/kratos)
