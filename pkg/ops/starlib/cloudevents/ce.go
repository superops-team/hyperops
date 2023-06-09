package cloudevents

import (
	"context"
	"fmt"
	"time"

	"encoding/base64"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	http "github.com/cloudevents/sdk-go/v2/protocol/http"
	localctx "github.com/superops-team/hyperops/pkg/ops/context"
	"github.com/superops-team/hyperops/pkg/ops/util"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const Name = "cloudevents"
const ModuleName = "cloudevents.star"

var Module = &starlarkstruct.Module{
	Name: "cloudevents",
	Members: starlark.StringDict{
		"report": localctx.AddBuiltin("cloudevents.report", Report),
	},
}

func Report(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var (
		addr    starlark.String
		auth    starlark.Tuple
		timeout starlark.Int
		headers = &starlark.Dict{}
		data    = &starlark.Dict{}
		ce      cloudevents.Client
		err     error
	)
	if err = starlark.UnpackArgs("cloudevents.report", args, kwargs, "addr", &addr, "headers", &headers, "data", &data, "auth?", &auth, "timeout?", &timeout); err != nil {
		return starlark.None, err
	}
	addrstr, err := util.AsString(addr)
	if err != nil {
		return starlark.None, err
	}
	if len(auth) == 2 {
		username, err := util.AsString(auth[0])
		if err != nil {
			return starlark.Bool(false), err
		}
		password, err := util.AsString(auth[1])
		if err != nil {
			return starlark.Bool(false), err
		}

		authstr := username + ":" + password
		basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(authstr))
		timeoutInt, ok := timeout.Int64()
		if !ok {
			timeoutInt = 5
		}
		ce, err = cloudevents.NewClientHTTP(
			http.WithTarget(addrstr),
			http.WithShutdownTimeout(time.Duration(timeoutInt)*time.Second),
			http.WithHeader("Authorization", basicAuth),
		)
		if err != nil {
			return starlark.None, err
		}
	} else {
		ce, err = cloudevents.NewClientHTTP(
			http.WithTarget(addrstr),
		)
		if err != nil {
			return starlark.None, err
		}
	}
	event := cloudevents.NewEvent()
	event.SetTime(time.Now())
	keys := headers.Keys()
	if len(keys) == 0 {
		return starlark.None, err
	}
	for _, key := range keys {
		keystr, err := util.AsString(key)
		if err != nil {
			continue
		}
		val, _, err := headers.Get(key)
		if err != nil {
			continue
		}
		if val.Type() != "string" {
			continue
		}
		valstr, err := util.AsString(val)
		if err != nil {
			continue
		}
		switch keystr {
		case "source":
			event.SetSource(valstr)
		case "type":
			event.SetType(valstr)
		default:
			event.SetExtension(keystr, valstr)
		}
	}
	cedata, err := util.Unmarshal(data)
	if err != nil {
		err = event.SetData(cloudevents.ApplicationJSON, map[string]string{"error": err.Error()})
		if err != nil {
			return starlark.None, err
		}
	} else {
		err = event.SetData(cloudevents.ApplicationJSON, cedata)
		if err != nil {
			return starlark.None, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	result := ce.Send(ctx, event)
	if cloudevents.IsNACK(result) {
		return starlark.None, fmt.Errorf("send event failed, %s %#v", result, cedata)
	}
	return starlark.Bool(true), nil
}
