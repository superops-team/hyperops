load("metric.star","metric")
load("time.star","time")

token = "b29ydC5tZXRyaXg6ZmVuZ3NodWFpdGFv"
domain = "http://metrix.byted.org"
query_instant = 'up{node="node06.hbcdcm02",service="kubelet"}'
query_range = 'up{node="node06.hbcdcm02",service="kubelet"}'

user = metric.new(token=token)

def show_instant():
    r = user.get_queries_by_instant(domain=domain,query=query_range,time=time.now())
    print(r.result_type)
    if r.result_type == "matrix":
        for i in r.result:
            for j in i.values:
                print(j)
            print(i.metric)
    elif r.result_type == "vector":
        for i in r.result:
            print("metric:",i.metric)
            print("value:",i.value)


def show_range():
    end_t = time.now()
    cost_t = time.second * 100
    start_t = end_t - cost_t
    r = user.get_queries_by_range(domain=domain,query=query_instant,step=10,start_time=start_t,end_time=end_t)
    print(r.result_type)
    if r.result_type == "matrix":
        for i in r.result:
            for j in i.values:
                print(j)
            print(i.metric)
    

def show_metadata_series():
    res = user.get_metadata(domain=domain,type="series",match=['up'])
    print("total series:",len(res))
    print("here is 5 examples:")
    for i in range(5):
        for j in res[i].keys():
            print(j,":",res[i][j])

def show_metadata_label():
    res = user.get_metadata(domain=domain,type="labels")
    print("total labels:",len(res))
    print("here is 5 examples:")
    for i in range(5):
        print(res[i])

def show_metadata_labelname():
    res = user.get_metadata(domain=domain,type="label_name",label_name="job")
    print("total label name:",len(res))
    print("here is 5 examples:")
    for i in range(5):
        print(res[i])

def show_rules():
    res = user.get_rules(domain=domain)
    print(res)

def show_targets():
    res = user.get_targets(domain=domain)
    for i in res.activeTargets:
        for j in i.keys():
            print(i,":",i[j])
    for i in res.droppedTargets:
        for j in i.keys():
            print(i,":",i[j])

def main():
    print("get query by instant :")
    show_instant()
    print("get query by range :")
    show_range()
    print("get metadata series :")
    show_metadata_series()
    print("get metadata labels :")
    show_metadata_label()
    print("get metadata label_name :")
    show_metadata_labelname()
    print("get rules :")
    #show_rules()
    print("get targets :")
    #show_targets()

main()