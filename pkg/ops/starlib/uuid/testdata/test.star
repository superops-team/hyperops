load("uuid.star", "uuid")
v3data = uuid.v3("elastic")
v4data = uuid.v4()
v5data = uuid.v5("elastic")
print(v3data)
print(v4data)
print(v5data)