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
