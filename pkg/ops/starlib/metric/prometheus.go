package metric

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/superops-team/hyperops/pkg/ops/util"
	starlarkTime "go.starlark.net/lib/time"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const StarlarkTimeFormat = "2006-01-02 15:04:05.999999999 -0700 MST"

var (
	ErrQueryArgs      = fmt.Errorf("usage metric.get_queries_by_instant(url,query,time)")
	ErrQueryRangeArgs = fmt.Errorf("usage metric.get_queries_by_range(url,query,start_time,end_time,step)")
	ErrMetaArgs       = fmt.Errorf("usage metric.get_metadata(url,type,match/label_name)")
	ErrData           = fmt.Errorf("can't parse data in response")
	ErrToken          = fmt.Errorf("token must set")
)

type UNIXTime float64

type Resp struct {
	Status    string      `json:"status"`
	ErrorType string      `json:"errorType"`
	Error     string      `json:"error"`
	Data      interface{} `json:"data"`
}

type queryResult struct {
	ResultType string `json:"resultType"`
}

type MatrixItem struct {
	Metric map[string]interface{} `json:"metric"`
	Values []interface{}          `json:"values"`
}
type Matrix struct {
	queryResult
	ResultList []MatrixItem `json:"result"`
}

//
type VectorItem struct {
	Metric map[string]interface{} `json:"metric"`
	Value  interface{}            `json:"value"`
}

type Vector struct {
	queryResult
	ResultList []VectorItem `json:"result"`
}

type Scalars interface{}

type metaData struct {
	Status string `json:"status"`
}
type Series struct {
	metaData
	Data []map[string]interface{} `json:"data"`
}

type LabelName struct {
	metaData
	Data []string `json:"data"`
}

type Label struct {
	metaData
	Data []string `json:"data"`
}

type ExemplarsItem struct {
	SeriesLabels map[string]string        `json:"seriesLabels"`
	Exemplars    []map[string]interface{} `json:"exemplars"`
}
type ExemplarsData struct {
	metaData
	Data []ExemplarsItem `json:"data"`
}

func Marshal(d interface{}) (starlark.Value, error) {
	//class data:
	//self.resultType = matrix | vector | scalar
	//self.result = []data
	var res starlark.Value
	switch data := d.(type) {
	case Matrix:
		var body []starlark.Tuple
		body = append(body, starlark.Tuple{starlark.String("result_type"), starlark.String("matrix")})
		var result starlark.Tuple
		for _, v := range data.ResultList {
			var kwarg []starlark.Tuple
			metric := starlark.NewDict(len(v.Metric))
			for key, _value := range v.Metric {
				_v, err := util.Marshal(_value)
				if err != nil {
					return starlark.None, err
				}
				_ = metric.SetKey(starlark.String(key), _v)
			}
			kwarg = append(kwarg, starlark.Tuple{starlark.String("metric"), metric})
			var value starlark.Tuple
			for _, item := range v.Values {
				res, err := parseItemValue(item)
				if err != nil {
					return starlark.None, err
				}
				value = append(value, res)
			}
			kwarg = append(kwarg, starlark.Tuple{starlark.String("values"), value})
			item := starlarkstruct.FromKeywords(starlark.String("data"), kwarg)
			result = append(result, item)
		}
		body = append(body, starlark.Tuple{starlark.String("result"), result})
		res = starlarkstruct.FromKeywords(starlark.String("matrix"), body)
	case Vector:
		var body []starlark.Tuple
		body = append(body, starlark.Tuple{starlark.String("result_type"), starlark.String("vector")})
		var result starlark.Tuple
		for _, v := range data.ResultList {
			var kwarg []starlark.Tuple
			metric := starlark.NewDict(len(v.Metric))
			for key, _value := range v.Metric {
				_v, err := util.Marshal(_value)
				if err != nil {
					return starlark.None, err
				}
				_ = metric.SetKey(starlark.String(key), _v)
			}
			kwarg = append(kwarg, starlark.Tuple{starlark.String("metric"), metric})
			//v.value应当是[]interface{}格式
			value, err := parseItemValue(v.Value)
			if err != nil {
				return starlark.None, err
			}
			kwarg = append(kwarg, starlark.Tuple{starlark.String("value"), value})
			item := starlarkstruct.FromKeywords(starlark.String("data"), kwarg)
			result = append(result, item)
		}
		body = append(body, starlark.Tuple{starlark.String("result"), result})
		res = starlarkstruct.FromKeywords(starlark.String("vector"), body)
	case Series:
		var tmp starlark.Tuple
		for _, dic := range data.Data {
			r, err := util.Marshal(dic)
			if err != nil {
				return starlark.None, err
			}
			tmp = append(tmp, r)
			res = tmp
		}
	case Label:
		var tmp starlark.Tuple
		for _, item := range data.Data {
			tmp = append(tmp, starlark.String(item))
		}
		res = tmp
	case LabelName:
		var tmp starlark.Tuple
		for _, item := range data.Data {
			tmp = append(tmp, starlark.String(item))
		}
		res = tmp
	case Scalars:
		//todo 没有测试代码覆盖到
		//比较低频率
		r, err := parseItemValue(data)
		if err != nil {
			return starlark.None, err
		}
		var body []starlark.Tuple
		body = append(body, starlark.Tuple{starlark.String("result_type"), starlark.String("scalars")})
		body = append(body, starlark.Tuple{starlark.String("result"), r})
		res = starlarkstruct.FromKeywords(starlark.String("scalars"), body)
	default:
		return starlark.None, ErrData
	}
	return res, nil
}

func NewRequest(method, url string, token string, body interface{}) (*http.Request, error) {
	buf := &bytes.Buffer{}
	if body != nil {
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, url, buf)
	req.Header.Set("Authorization", "Basic "+token)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func DoRequest(req *http.Request, v interface{}, timeout int) error {
	if timeout <= 0 || timeout > 60 {
		timeout = 10
	}
	client := http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if status := resp.StatusCode; status < 200 || status > 299 {
		res, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("status code %d - %s", status, string(res))
	}
	if v != nil {
		err = json.NewDecoder(resp.Body).Decode(v)
	}
	return err
}

func GetUrl(urlHead, method string, param map[string]interface{}) (string, error) {
	var domain string
	switch method {
	case "query":
		domain = urlHead + "/api/v1/query"
	case "query_range":
		domain = urlHead + "/api/v1/query_range"
	default:
		return "", fmt.Errorf("got unexpected method")
	}
	value := url.Values{}
	for k, v := range param {
		switch _v := v.(type) {
		case int64:
			value.Add(k, fmt.Sprintf("%d", _v))
		case int:
			value.Add(k, fmt.Sprintf("%d", _v))
		case string:
			value.Add(k, _v)
		case float64:
			value.Add(k, fmt.Sprintf("%f", _v))
		}
	}
	return domain + "?" + value.Encode(), nil
}

func parseItemValue(v interface{}) (starlark.Value, error) {
	tem, ok := v.([]interface{})
	if !ok {
		return starlark.None, ErrData
	}
	if len(tem) != 2 {
		return starlark.None, ErrData
	}
	duration, ok := tem[0].(float64)
	if !ok {
		return starlark.None, ErrData
	}
	_v, err := util.Marshal(tem[1])
	if err != nil {
		return starlark.None, err
	}
	value := starlark.Tuple{starlark.Float(duration), _v}
	return value, nil
}

func getTimeString(t starlarkTime.Time) string {
	res := t.String()
	idx := strings.IndexAny(res, "m=")
	if idx == -1 {
		return res
	}
	return res[:idx-1]
}
