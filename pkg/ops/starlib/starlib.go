package starlib

import (
	"fmt"

	"github.com/superops-team/hyperops/pkg/ops/starlib/cloudevents"
	"github.com/superops-team/hyperops/pkg/ops/starlib/compress/gzip"
	"github.com/superops-team/hyperops/pkg/ops/starlib/encoding/base64"
	"github.com/superops-team/hyperops/pkg/ops/starlib/encoding/csv"
	"github.com/superops-team/hyperops/pkg/ops/starlib/encoding/json"
	"github.com/superops-team/hyperops/pkg/ops/starlib/encoding/yaml"
	"github.com/superops-team/hyperops/pkg/ops/starlib/env"
	"github.com/superops-team/hyperops/pkg/ops/starlib/fs"
	"github.com/superops-team/hyperops/pkg/ops/starlib/group"
	"github.com/superops-team/hyperops/pkg/ops/starlib/hash"
	"github.com/superops-team/hyperops/pkg/ops/starlib/http"
	"github.com/superops-team/hyperops/pkg/ops/starlib/math"
	"github.com/superops-team/hyperops/pkg/ops/starlib/re"
	"github.com/superops-team/hyperops/pkg/ops/starlib/sh"
	"github.com/superops-team/hyperops/pkg/ops/starlib/sys"
	"github.com/superops-team/hyperops/pkg/ops/starlib/time"
	"github.com/superops-team/hyperops/pkg/ops/starlib/tools"
	"github.com/superops-team/hyperops/pkg/ops/starlib/uuid"
	"github.com/superops-team/hyperops/pkg/ops/starlib/zipfile"
	"go.starlark.net/starlark"
)

const Version = "0.1.0"

// Loader presents the starlib library as a loader
func Loader(thread *starlark.Thread, module string) (dict starlark.StringDict, err error) {
	switch module {
	case time.ModuleName:
		return starlark.StringDict{"time": time.Module}, nil
	case gzip.ModuleName:
		return starlark.StringDict{"gzip": gzip.Module}, nil
	case http.ModuleName:
		return http.LoadModule()
	case re.ModuleName:
		return re.LoadModule()
	case base64.ModuleName:
		return base64.LoadModule()
	case csv.ModuleName:
		return csv.LoadModule()
	case json.ModuleName:
		return starlark.StringDict{"json": json.Module}, nil
	case yaml.ModuleName:
		return yaml.LoadModule()
	case math.ModuleName:
		return starlark.StringDict{"math": math.Module}, nil
	case hash.ModuleName:
		return hash.LoadModule()
	case uuid.ModuleName:
		return starlark.StringDict{"uuid": uuid.Module}, nil
	case zipfile.ModuleName:
		return starlark.StringDict{"zipfile": zipfile.Module}, nil
	case group.ModuleName:
		return starlark.StringDict{"group": group.Module}, nil
	case sh.ModuleName:
		return starlark.StringDict{"shell": sh.Module}, nil
	case env.ModuleName:
		return starlark.StringDict{"env": env.Module}, nil
	case sys.ModuleName:
		return starlark.StringDict{"sys": sys.Module}, nil
	case fs.ModuleName:
		return starlark.StringDict{"fs": fs.Module}, nil
	case tools.ModuleName:
		return starlark.StringDict{"tools": tools.Module}, nil
	case cloudevents.ModuleName:
		return starlark.StringDict{"cloudevents": cloudevents.Module}, nil
	}

	return nil, fmt.Errorf("invalid module %q", module)
}
