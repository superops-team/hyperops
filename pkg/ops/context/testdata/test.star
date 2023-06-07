# predeclared globals for test: ctx
load("assert.star", "assert")

assert.eq(ctx.get_config("foo"), "bar")
assert.eq(ctx.get_secret("baz"), "bat")

ctx.set("foo", "bar")
assert.eq(ctx.get("foo"), "bar")
ctx.set_secret("pass", "12345")
print(ctx.get_secret("pass"))
