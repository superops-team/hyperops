
print("hello world")

# output(sys.argv)
ret = sh(dir="./", cmd="ls -a", timeout=10)

print(ret.stdout)
print(ctx.get_secret("password1"))
print(ctx.get_config("x"))
