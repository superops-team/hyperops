load('localcache.star', 'localcache')
load('assert.star', 'assert')

lc = localcache.new()

lc.set("test1", "val1")
lc.set("test2", "val2")
lc.set("test3", "val3")
lc.set("test4", "val4")
lc.set("test5", "val5")

exist_flag = lc.exist("test1")
assert.eq(exist_flag, True)

val = lc.get("test1")
assert.eq(val, "val1")
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
    assert.eq(exist_flag1, True)

    val = lcm.get("test1")
    assert.eq(val, "val1")
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
