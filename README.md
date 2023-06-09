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

* shell模块: 执行shell系统调用

```
load('shell.star', 'shell')

res_1 = shell.exec(dir="./", cmd="ls -l", timeout=5)
print(res_1.code)
print(res_1.stdout)
print(res_1.stderr)

res_3 = shell.exec("df -lh")
print(res_3)
res_2 = shell.exec(dir="./", cmd="pwd")
print(res_2)

res_5 = shell.exec(dir="./", cmd="pwdss")
print(res_5)

```


* fs模块：文件的各种操作

```
load('fs.star', 'fs')

fs.create("testdata/test.txt", "1234")

datafss = fs.ls()
print(datafss)

fss = fs.glob("./*")
print(fss)

eflag  = fs.exist("testdata/test.txt")
print(eflag)

content = fs.readall("testdata/test.txt")
print(content)

appendflag = fs.append("testdata/test.txt", "123")
print(appendflag)

writeflag = fs.create("testdata/test.txt", "hello,world")
print(writeflag)

stat = fs.stat("testdata/test.txt")
print(stat)

dirname = fs.dirname("testdata/test.txt")
print(dirname)

basename = fs.basename("testdata/test.txt")
print(basename)

md5sum = fs.md5("testdata/test.txt")
print(md5sum)

compressflag = fs.gzip("testdata/test.txt", "testdata/test.txt.tar.gz")
print(compressflag)

flag = fs.rm("testdata/test.txt")
print(flag)
rflag = fs.rm("testdata/test.txt.tar.gz")
print(rflag)
```

* localcache模块：基于LSM tree实现一个本地存储的模型，支持快速存储执行历史和记录上下文

```
load('localcache.star', 'localcache')

lc = localcache.new()

lc.set("test1", "val1")
lc.set("test2", "val2")
lc.set("test3", "val3")
lc.set("test4", "val4")
lc.set("test5", "val5")

exist_flag = lc.exist("test1")

val = lc.get("test1")
print(val)

def test_filter():
    data = lc.filter("test")
    for k, v in data.items():
        print(k, v)

test_filter()

def test_filter_key():
    data = lc.filter_key("test")
    for v in data:
        print(v)

test_filter_key()
lc.delete("test3")
lc.clear("test1")

def test_in_memory():
    lcm = localcache.new()

    lcm.set("test1", "val1")
    lcm.set("test2", "val2")
    lcm.set("test3", "val3")
    lcm.set("test4", "val4")
    lcm.set("test5", "val5")

    exist_flag1 = lcm.exist("test1")

    val = lcm.get("test1")
    print(val)

    def test_filter():
        data = lcm.filter("test")
        for k, v in data.items():
            print(k, v)

    test_filter()

    def test_filter_key():
        data = lcm.filter_key("test")
        for v in data:
            print(v)

    test_filter_key()
    lcm.delete("test3")
    lcm.clear("test1")

test_in_memory()
```
