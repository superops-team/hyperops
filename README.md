<p align="center">
    <img src="https://github.com/superops-team/hyperops/blob/main/docs/hyperops-overview.png" class="center" />
</p>

### HyperOps

> hyperops is an easy-to-use, cloud-native, cross-platform, operations-oriented programming language.

----

<p align="center">
    <a href="https://godoc.org/github.com/superops-team/hyperops"><img src="https://godoc.org/github.com/superops-team/hyperops?status.svg"/></a>
    <a href="https://github.com/superops-team/hyperops"><img src="https://img.shields.io/badge/release-v0.1.3-blue"/></a>
    <a href="https://goreportcard.com/report/github.com/superops-team/hyperops"><img src="https://goreportcard.com/badge/github.com/superops-team/hyperops"/></a>
    <a href="https://github.com/avelino/awesome-go"><img src="https://awesome.re/mentioned-badge.svg"/></a>
</p>

## Intro

Hyperops is a cloudnative ops language for better ops script.

## Slogan

```
Better Ops For Better Life.
```

## Features

* Measurable

The metrics was inside kernel and runtime, you can easily known your ops script with metrics show diagrams on grafana.
The log was designed by ops domain with easily transfer to any log platform.
The script can be trigger any event by the hook in the kernel.

* Write once run any where

The kernel was writter in golang with better performance and multiplatform supports.
The same ops script can be run different platform like x86, arm, mac, windows etc.

* Easy to use

The most grammers was came from python, and it came from starlark.
The libs in ops domain was inside the kernel.


## How to install

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

## How to use

* Hello,world

write a script named `hello.ops`

```
print("hello,world")
```

run it

```
hyperops apply -f hello.ops
```
