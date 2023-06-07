load('cloudevents.star', 'cloudevents')

headers = {
    "source": "test/123",
    "type": "hyperops/event",
    "extProduct": "hyperops",
    "extPSM": "hyperops",
    "extEnvironment": "dev",
    "extIP": "127.0.0.1",
}

data = {
    "test": "123",
}

addr = "http://defensor-boe.byted.org/api/v1/event/cloud"

ret1 = cloudevents.report(addr=addr, headers=headers, data=data, timeout=2)
print(ret1)
ret2 = cloudevents.report(addr=addr, headers=headers, data=data, timeout=2, auth=("cloudevent", "password"))
print(ret2)
