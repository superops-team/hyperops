package metric

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/superops-team/hyperops/pkg/ops/context"
	"github.com/superops-team/hyperops/pkg/ops/util"
	"github.com/mitchellh/mapstructure"
	starlarkTime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const Name = "metric"
const ModuleName = "metric.star"

var Module = &starlarkstruct.Module{
	Name: "metric",
	Members: starlark.StringDict{
		"new": context.AddBuiltin("metric.new", New),
	},
}

type Metric struct {
	token string
}

func New(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	token, err := params.GetStringByName("token")
	if err != nil {
		token, err = params.GetString(0)
		if err != nil {
			return starlark.None, err
		}
	}
	t := Metric{
		token: token,
	}
	sm := context.NewSecretsManager()
	sm.AddSecret(t.token)
	return t.Struct(), nil
}

func (t *Metric) Struct() *starlarkstruct.Struct {
	return starlarkstruct.FromStringDict(
		starlark.String("metric"),
		starlark.StringDict{
			"get_queries_by_instant": context.AddBuiltin("metric.get_queries", t.GetQueriesByInstant),
			"get_queries_by_range":   context.AddBuiltin("metric.get_queries_by_range", t.GetQueriesByRange),
			"get_metadata":           context.AddBuiltin("metric.get_metadata", t.GetMetaData),
			"get_rules":              context.AddBuiltin("metric.get_rules", t.GetRules),
			"get_targets":            context.AddBuiltin("metric.get_targets", t.GetTarget),
		},
	)
}

func (metric *Metric) GetQueriesByRange(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	//usage metric.get_queries_by_range(url="",query="",start_time="",end_time="",step=int)
	var (
		endTime   time.Time
		startTime time.Time
		query     string
		step      int64
		domain    string
		res       starlark.Value
	)
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	domain, err = params.GetStringByName("domain")
	if err != nil {
		domain, err = params.GetString(0)
		if err != nil {
			return starlark.None, ErrQueryRangeArgs
		}
	}
	query, err = params.GetStringByName("query")
	if err != nil {
		query, err = params.GetString(1)
		if err != nil {
			return starlark.None, ErrQueryRangeArgs
		}
	}
	var idx int
	idx, err = params.GetParamIndex("start_time")
	if err != nil {
		idx = 2
	}
	v, err := params.GetParam(idx)
	if err != nil {
		return starlark.None, ErrQueryRangeArgs
	}
	startV, ok := v.(starlarkTime.Time)
	if !ok {
		return starlark.None, ErrQueryRangeArgs
	}
	if err != nil {
		return starlark.None, err
	}
	startTime, err = time.Parse(StarlarkTimeFormat, getTimeString(startV))
	if err != nil {
		return starlark.None, err
	}
	idx, err = params.GetParamIndex("end_time")
	if err != nil {
		idx = 3
	}
	v, err = params.GetParam(idx)
	if err != nil {
		return starlark.None, ErrQueryRangeArgs
	}
	endV, ok := v.(starlarkTime.Time)
	if !ok {
		return starlark.None, ErrQueryArgs
	}
	endTime, err = time.Parse(StarlarkTimeFormat, getTimeString(endV))
	if err != nil {
		return starlark.None, err
	}
	step, err = params.GetIntByName("step")
	if err != nil {
		step, err = params.GetInt(4)
		if err != nil {
			step = 10
		}
	}
	timeout, err := params.GetIntByName("timeout")
	if err != nil {
		timeout, err = params.GetInt(5)
		if err != nil {
			timeout = 10
		}
	}

	dict := make(map[string]interface{})
	dict["end"] = endTime.Format(time.RFC3339)
	dict["start"] = startTime.Format(time.RFC3339)
	dict["step"] = step
	dict["query"] = query
	domain, err = GetUrl(domain, "query_range", dict)
	if err != nil {
		return starlark.None, err
	}
	resp, err := NewRequest(http.MethodGet, domain, metric.token, nil)
	if err != nil {
		return starlark.None, err
	}
	var Info Resp
	err = DoRequest(resp, &Info, int(timeout))
	if err != nil {
		return starlark.None, err
	}
	data, ok := Info.Data.(map[string]interface{})
	if !ok {
		return starlark.None, ErrData
	}
	resultType, ok := data["resultType"]
	if !ok {
		return starlark.None, ErrData
	}
	switch resultType {
	case "matrix":
		v := Matrix{}
		t, ok := data["result"]
		if !ok {
			return starlark.None, ErrData
		}
		err := mapstructure.Decode(t, &v.ResultList)
		if err != nil {
			return starlark.None, err
		}
		res, err = Marshal(v)
		if err != nil {
			return starlark.None, err
		}
	case "vector":
		v := Vector{}
		t, ok := data["result"]
		if !ok {
			return starlark.None, ErrData
		}
		err := mapstructure.Decode(t, &v.ResultList)
		if err != nil {
			return starlark.None, err
		}
		res, err = Marshal(v)
		if err != nil {
			return starlark.None, err
		}
	case "scalar":
		var v Scalars
		t, ok := data["result"]
		if !ok {
			return starlark.None, ErrData
		}
		err := mapstructure.Decode(t, &v)
		if err != nil {
			return starlark.None, err
		}
		res, err = Marshal(v)
		if err != nil {
			return starlark.None, err
		}
	}
	return res, nil
}
func (metric *Metric) GetQueriesByInstant(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	//usage metric.get_queries_by_instant(url="",query="",time="")
	var (
		queryTime time.Time
		query     string
		domain    string
		res       starlark.Value
	)
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	domain, err = params.GetStringByName("domain")
	if err != nil {
		domain, err = params.GetString(0)
		if err != nil {
			return starlark.None, ErrQueryArgs
		}
	}
	query, err = params.GetStringByName("query")
	if err != nil {
		query, err = params.GetString(1)
		if err != nil {
			return starlark.None, ErrQueryArgs
		}
	}

	t, err := params.GetStringByName("time")
	if err != nil {
		t, err = params.GetString(2)
		if err != nil {
			t = time.Now().Format(time.RFC3339)
		}
	}
	queryTime, err = time.Parse(time.RFC3339, t)
	if err != nil {
		return starlark.None, err
	}
	timeout, err := params.GetIntByName("timeout")
	if err != nil {
		timeout, err = params.GetInt(3)
		if err != nil {
			timeout = 10
		}
	}

	dict := make(map[string]interface{})
	dict["query"] = query
	dict["time"] = queryTime.Format(time.RFC3339)
	domain, err = GetUrl(domain, "query", dict)
	if err != nil {
		return starlark.None, err
	}
	resp, err := NewRequest(http.MethodGet, domain, metric.token, nil)
	if err != nil {
		return starlark.None, err
	}
	var Info Resp
	err = DoRequest(resp, &Info, int(timeout))
	if err != nil {
		return starlark.None, err
	}
	data, ok := Info.Data.(map[string]interface{})
	if !ok {
		return starlark.None, ErrData
	}
	resultType, ok := data["resultType"]
	if !ok {
		return starlark.None, ErrData
	}
	switch resultType {
	case "matrix":
		v := Matrix{}
		t, ok := data["result"]
		if !ok {
			return starlark.None, ErrData
		}
		err := mapstructure.Decode(t, &v.ResultList)
		if err != nil {
			return starlark.None, err
		}
		res, err = Marshal(v)
		if err != nil {
			return starlark.None, err
		}
	case "vector":
		v := Vector{}
		t, ok := data["result"]
		if !ok {
			return starlark.None, ErrData
		}
		err := mapstructure.Decode(t, &v.ResultList)
		if err != nil {
			return starlark.None, err
		}
		res, err = Marshal(v)
		if err != nil {
			return starlark.None, err
		}
	case "scalar":
		var v Scalars
		t, ok := data["result"]
		if !ok {
			return starlark.None, ErrData
		}
		err := mapstructure.Decode(t, &v)
		if err != nil {
			return starlark.None, err
		}
		res, err = Marshal(v)
		if err != nil {
			return starlark.None, err
		}
	}

	return res, nil
}

func (metric *Metric) GetMetaData(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	//finding series by label matchers :type=series
	//getting label names :type=label_name
	//querying label values:type=label
	//usage matric.get_metadata(url="",type="",match="",label_name="")
	//match is a list
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		if err != nil {
			return starlark.None, ErrMetaArgs
		}
	}
	domain, err := params.GetStringByName("domain")
	if err != nil {
		domain, err = params.GetString(0)
		if err != nil {
			return starlark.None, ErrMetaArgs
		}
	}
	dataType, err := params.GetStringByName("type")
	if err != nil {
		dataType, err = params.GetString(1)
		if err != nil {
			return starlark.None, ErrMetaArgs
		}
	}
	var match []string
	for _, kwarg := range kwargs {
		if kwarg[0] == starlark.String("match") {
			matchList := kwarg[1]
			switch lis := matchList.(type) {
			case starlark.Tuple:
				for i := 0; i < lis.Len(); i++ {
					v := lis.Index(i)
					stringv, ok := v.(starlark.String)
					if !ok {
						return starlark.None, ErrData
					}
					match = append(match, stringv.GoString())
				}
			case *starlark.List:
				for i := 0; i < lis.Len(); i++ {
					v := lis.Index(i)
					stringv, ok := v.(starlark.String)
					if !ok {
						return starlark.None, ErrData
					}
					match = append(match, stringv.GoString())
				}
			default:
				return starlark.None, fmt.Errorf("usage a list or tuple as the value of match")
			}
			break
		}
	}
	start, err := params.GetStringByName("start_time")
	if err != nil {
		start = "none"
	}
	end, err := params.GetStringByName("end_time")
	if err != nil {
		end = "none"
	}
	label_name, err := params.GetStringByName("label_name")
	if err != nil {
		label_name = "none"
	}
	timeout, err := params.GetIntByName("timeout")
	if err != nil {
		timeout = 10
	}

	value := url.Values{}
	switch dataType {
	case "series":
		//series 只需要match start end三个参数 而match是必须的
		domain = domain + "/api/v1/series"
		if len(match) == 0 {
			return starlark.None, fmt.Errorf("match must be set")
		}
		for _, v := range match {
			value.Add("match[]", v)
		}
	case "labels":
		domain = domain + "/api/v1/labels"
	case "label_name":
		if label_name == "none" {
			return starlark.None, fmt.Errorf("label_name must be set")
		}
		domain = domain + fmt.Sprintf("/api/v1/label/%s/values", label_name)
	default:
		return starlark.None, fmt.Errorf("matadata type: series,labels,label_name")
	}
	if start != "none" {
		value.Add("start", start)
	}
	if end != "none" {
		value.Add("end", end)
	}
	domain = domain + "?" + value.Encode()
	req, err := NewRequest(http.MethodGet, domain, metric.token, nil)
	if err != nil {
		return starlark.None, err
	}
	var Info interface{}
	err = DoRequest(req, &Info, int(timeout))
	if err != nil {
		return starlark.None, err
	}
	var res starlark.Value
	data, ok := Info.(map[string]interface{})
	if !ok {
		return starlark.None, ErrData
	}
	switch dataType {
	case "series":
		v := Series{}
		t, ok := data["data"]
		if !ok {
			return starlark.None, ErrData
		}
		err := mapstructure.Decode(t, &v.Data)
		if err != nil {
			return starlark.None, err
		}
		res, err = Marshal(v)
		if err != nil {
			return starlark.None, err
		}
	case "labels":
		v := Label{}
		t, ok := data["data"]
		if !ok {
			return starlark.None, ErrData
		}
		err := mapstructure.Decode(t, &v.Data)
		if err != nil {
			return starlark.None, err
		}
		res, err = Marshal(v)
		if err != nil {
			return starlark.None, err
		}
	case "label_name":
		v := LabelName{}
		t, ok := data["data"]
		if !ok {
			return starlark.None, ErrData
		}
		err := mapstructure.Decode(t, &v.Data)
		if err != nil {
			return starlark.None, err
		}
		res, err = Marshal(v)
		if err != nil {
			return starlark.None, err
		}
	}
	return res, nil
}

func (metric *Metric) GetRules(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var res starlark.Value
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	domain, err := params.GetStringByName("domain")
	if err != nil {
		domain, err = params.GetString(0)
		if err != nil {
			return starlark.None, err
		}
	}
	domain = domain + "/api/v1/rules"
	timeout, err := params.GetIntByName("timeout")
	if err != nil {
		timeout, err = params.GetInt(1)
		if err != nil {
			timeout = 10
		}
	}

	resp, err := NewRequest(http.MethodGet, domain, metric.token, nil)
	if err != nil {
		return starlark.None, err
	}

	var Info interface{}
	err = DoRequest(resp, &Info, int(timeout))
	if err != nil {
		return starlark.None, err
	}
	data, ok := Info.(map[string]interface{})
	if !ok {
		return starlark.None, ErrData
	}
	_data, ok := data["data"]
	if !ok {
		return starlark.None, ErrData
	}
	dataDict, ok := _data.(map[string]interface{})
	if !ok {
		return starlark.None, ErrData
	}
	group, ok := dataDict["groups"]
	if !ok {
		return starlark.None, ErrData
	}
	tuple := starlark.Tuple{}
	groups, ok := group.([]map[string]interface{})
	if !ok {
		if emptyV, ok := group.([]interface{}); ok {
			res, err = util.Marshal(emptyV)
			if err != nil {
				return starlark.None, err
			}
			return res, nil
		}
		return starlark.None, ErrData
	}
	for _, v := range groups {
		tmp, err := util.Marshal(v)
		if err != nil {
			return starlark.None, err
		}
		tuple = append(tuple, tmp)
	}
	res = tuple
	return res, nil
}

func (metric *Metric) GetTarget(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var res starlark.Value
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	domain, err := params.GetStringByName("domain")
	if err != nil {
		domain, err = params.GetString(0)
		if err != nil {
			return starlark.None, err
		}
	}
	domain = domain + "/api/v1/targets"
	timeout, err := params.GetIntByName("timeout")
	if err != nil {
		timeout, err = params.GetInt(1)
		if err != nil {
			timeout = 10
		}
	}

	resp, err := NewRequest(http.MethodGet, domain, metric.token, nil)
	if err != nil {
		return starlark.None, err
	}
	var Info interface{}
	err = DoRequest(resp, &Info, int(timeout))
	if err != nil {
		return starlark.None, err
	}
	data, ok := Info.(map[string]interface{})
	if !ok {
		return starlark.None, ErrData
	}
	var kwarg []starlark.Tuple
	targetV, ok := data["data"]
	if !ok {
		return starlark.None, ErrData
	}
	target, ok := targetV.(map[string]interface{})
	if !ok {
		return starlark.None, ErrData
	}
	discoverd, ok := target["activeTargets"]
	if !ok {
		return starlark.None, ErrData
	}
	discoverdV, err := util.Marshal(discoverd)
	if err != nil {
		return starlark.None, ErrData
	}
	kwarg = append(kwarg, starlark.Tuple{starlark.String("activeTargets"), discoverdV})
	dropped, ok := target["droppedTargets"]
	if !ok {
		return starlark.None, ErrData
	}
	droppedV, err := util.Marshal(dropped)
	if err != nil {
		return starlark.None, ErrData
	}
	kwarg = append(kwarg, starlark.Tuple{starlark.String("droppedTargets"), droppedV})
	res = starlarkstruct.FromKeywords(starlark.String("Targets"), kwarg)
	return res, nil
}
