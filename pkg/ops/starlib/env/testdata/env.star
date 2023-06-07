load('env.star', 'env')

env.set(key="byte", val="dance")
env.set("hello", "world")
res1 = env.get(key="byte")
res2 = env.get("hello")
print(res1)
print(res2)
