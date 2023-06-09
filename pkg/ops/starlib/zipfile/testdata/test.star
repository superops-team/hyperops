load("assert.star", "assert")
load("zipfile.star", "zipfile")

zr = zipfile.new(hello_world_zip)
assert.eq(zr.namelist(), ["testdata/", "testdata/world/","testdata/world/world.txt","testdata/hello.txt"])
assert.eq(zr.open("testdata/hello.txt").read(), "hello\n")
