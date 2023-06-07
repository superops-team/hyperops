load('file.star', 'file')

datafiles = file.ls("/data/tiger/edge_collector", 10)
print(datafiles)

files = file.glob("/data/*/*")
print(files)


f = file.open("testdata/test.txt", "wr")
f.write("hahahah")
f.close()

fs = file.exist("testdata/test.txt")
print(fs)

content = file.readall("testdata/test.txt")
print(content)

appendflag = file.append("testdata/test.txt", "123")
print(appendflag)

writeflag = file.create("testdata/test.txt", "hello,world")
print(writeflag)

stat = file.stat("testdata/test.txt")
print(stat)

dirname = file.dirname("testdata/test.txt")
print(dirname)

basename = file.basename("testdata/test.txt")
print(basename)

md5sum = file.md5("testdata/test.txt")
print(md5sum)

compressflag = file.gzip("testdata/test.txt", "testdata/test.txt.tar.gz")
print(compressflag)

#flag = file.rm("testdata/test.txt")
#print(flag)
