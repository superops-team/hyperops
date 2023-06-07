load("sys.star", "sys")
load('assert.star', 'assert')


print(sys.os)
print(sys.arch)
print(sys.platform)

assert.eq(sys.platform, sys.os + "_" + sys.arch)
