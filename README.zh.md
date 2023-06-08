# Hyperops介绍

Hyperops 是面向云原生的运维自动化编程语言，其旨在简化运维编程，通过提供通用型，易于编写，高性能，可度量的运维脚本来取代shell script 和 大多数需要用yaml编排的运维场景.

## 背景介绍

我编写Hyperops的宗旨是简化运维成本，在运维自动化领域提供一个更简单有效，并且可靠的运行时来简化工作，我非常喜欢python编程语言的灵活，但是使用python做跨平台我非常苦恼，我非常喜欢go语言的高效和简介以及其内置的强类型提供的更高的安全性，我非常不喜欢也不擅长使用shell来编写脚本，每次编写shell我都会求助于Google来获取shell中非常不常见的定制化编程特性，这使得我每次编写shell都非常苦恼，我特别希望有一个面向运维场景的编程语言来解决我所面临的问题，但是很遗憾并没有一个好用的编程语言能够具备我希望拥有的运维领域的特性，因此我发起了Hyperops项目。


## Hyperops特性

1. 执行引擎内置可度量指标，日志，事件

2. 跨平台，编写一次多平台可执行

3. 易于使用，大部分语法来自python，无需重新学习新的编程语言

4. 内置状态管理模型，支持持久化执行过程，执行历史

5. 内置安全鉴权管理控制接入功能，方便的执行审计和鉴权


## 如何安装

```
# step1: clone the project
git clone https://github.com/superops-team/hyperops.git

# step2: change workdir
cd hyperops

# step3: pull deps
make deps

# step4: build hyperops into binary
make build

# step6: success to use the hyperops tool
./bin/hyperops
```

## 如何使用

```bash
./bin/hyperops apply -f examples/hello_world.ops
```

## 更多内置包，功能介绍

TODO
