package fs

import (
	"archive/tar"
	"compress/gzip"
	"crypto/md5"
	"fmt"
	"syscall"
	"time"

	"io"
	"io/ioutil"
	"path/filepath"

	"os"

	localctx "github.com/superops-team/hyperops/pkg/ops/context"
	"github.com/superops-team/hyperops/pkg/ops/util"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const Name = "fs"
const ModuleName = "fs.star"

var Module = &starlarkstruct.Module{
	Name: "fs",
	Members: starlark.StringDict{
		"readall":  localctx.AddBuiltin("fs.readall", ReadAll),
		"create":   localctx.AddBuiltin("fs.create", Create),
		"append":   localctx.AddBuiltin("fs.append", Append),
		"md5":      localctx.AddBuiltin("fs.md5", Md5),
		"gzip":     localctx.AddBuiltin("fs.gzip", Gzip),
		"exist":    localctx.AddBuiltin("fs.exist", Exist),
		"stat":     localctx.AddBuiltin("fs.stat", Stat),
		"glob":     localctx.AddBuiltin("fs.glob", Glob),
		"ls":       localctx.AddBuiltin("fs.ls", Ls),
		"basename": localctx.AddBuiltin("fs.basename", Basename),
		"dirname":  localctx.AddBuiltin("fs.dirname", Dirname),
		"rm":       localctx.AddBuiltin("fs.rm", Remove),
	},
}

func Gzip(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	filename, err := params.GetString(0)
	if err != nil {
		filename, err = params.GetStringByName("file")
		if err != nil {
			return starlark.None, err
		}
	}
	output, err := params.GetString(1)
	if err != nil {
		output, err = params.GetStringByName("output")
		if err != nil {
			return starlark.None, err
		}
	}
	file, err := os.Open(filename)
	if err != nil {
		return starlark.Bool(false), err
	}
	defer file.Close()

	// Create the output file
	outputfile, err := os.Create(output)
	if err != nil {
		return starlark.Bool(false), err
	}
	defer outputfile.Close()
	// Create a gzip writer
	gzipWriter := gzip.NewWriter(outputfile)
	defer gzipWriter.Close()

	// Create a tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Write the file to the tar archive
	fileInfo, err := file.Stat()
	if err != nil {
		return starlark.Bool(false), err
	}

	header := &tar.Header{
		Name: fileInfo.Name(),
		Size: fileInfo.Size(),
		Mode: int64(fileInfo.Mode()),
	}

	if err := tarWriter.WriteHeader(header); err != nil {
		return starlark.Bool(false), err
	}

	if _, err := io.Copy(tarWriter, file); err != nil {
		return starlark.Bool(false), err
	}
	return starlark.Bool(true), nil
}

func Md5(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	filename, err := params.GetString(0)
	if err != nil {
		filename, err = params.GetStringByName("filepath")
		if err != nil {
			return starlark.None, err
		}
	}
	file, err := os.Open(filename)
	if err != nil {
		return starlark.None, err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return starlark.None, err
	}
	return starlark.String(fmt.Sprintf("%x", hash.Sum(nil))), nil
}

func Dirname(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	filename, err := params.GetString(0)
	if err != nil {
		filename, err = params.GetStringByName("filepath")
		if err != nil {
			return starlark.None, err
		}
	}
	dirname := filepath.Dir(filename)
	return starlark.String(dirname), nil
}

func Basename(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	filename, err := params.GetString(0)
	if err != nil {
		filename, err = params.GetStringByName("filepath")
		if err != nil {
			return starlark.None, err
		}
	}
	basename := filepath.Base(filename)
	return starlark.String(basename), nil
}

func Remove(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.Bool(false), err
	}
	filePath, err := params.GetString(0)
	if err != nil {
		filePath, err = params.GetStringByName("filepath")
		if err != nil {
			return starlark.Bool(false), err
		}
	}
	flag, err := params.GetString(1)
	if err != nil {
		flag, err = params.GetStringByName("args")
		if err != nil {
			flag = ""
		}
	}
	if flag == "all" {
		err = os.RemoveAll(filePath)
	} else {
		err = os.Remove(filePath)
	}
	if err != nil {
		return starlark.Bool(false), err
	}
	return starlark.Bool(true), nil
}

func Ls(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	dirPath, err := params.GetString(0)
	if err != nil {
		dirPath, err = params.GetStringByName("dir")
		if err != nil {
			dirPath = "./"
		}
	}

	num, err := params.GetInt(1)
	if err != nil {
		num, err = params.GetIntByName("limit")
		if err != nil {
			num = -1
		}
	}
	// 打开目录
	dir, err := os.Open(dirPath)
	if err != nil {
		return starlark.None, err
	}
	defer dir.Close()

	// 读取目录中的文件
	files, err := dir.Readdir(int(num))
	if err != nil {
		return starlark.None, err
	}

	// 遍历目录中的文件
	fileList := []interface{}{}
	for _, file := range files {
		item := convertFileInfo(file)
		fileList = append(fileList, item)
	}
	return util.ConvertToStarlark(fileList)
}

func convertFileInfo(fileStat os.FileInfo) starlark.Value {
	sysStat := fileStat.Sys().(*syscall.Stat_t)

	atime := time.Unix(int64(sysStat.Atim.Sec), int64(sysStat.Atim.Nsec))
	mtime := time.Unix(int64(sysStat.Mtim.Sec), int64(sysStat.Mtim.Nsec))
	ctime := time.Unix(int64(sysStat.Ctim.Sec), int64(sysStat.Ctim.Nsec))

	return starlarkstruct.FromStringDict(starlarkstruct.Default, starlark.StringDict{
		"name":     starlark.String(fileStat.Name()),
		"size":     starlark.MakeInt64(fileStat.Size()),
		"mode":     starlark.String(fileStat.Mode().String()),
		"modtime":  starlark.String(fileStat.ModTime().Format("2006-01-02 15:04:05")),
		"isdir":    starlark.Bool(fileStat.IsDir()),
		"dev":      starlark.MakeUint64(sysStat.Dev),
		"inode":    starlark.MakeUint64(sysStat.Ino),
		"mode_int": starlark.MakeUint64(uint64(sysStat.Mode)),
		"nlink":    starlark.MakeUint64(uint64(sysStat.Nlink)),
		"uid":      starlark.MakeUint64(uint64(sysStat.Uid)),
		"gid":      starlark.MakeUint64(uint64(sysStat.Gid)),
		"rdev":     starlark.MakeUint64(uint64(sysStat.Mode)),
		"blksize":  starlark.MakeUint64(uint64(sysStat.Blksize)),
		"blocks":   starlark.MakeInt64(sysStat.Blocks),
		"atime":    starlark.String(atime.Format("2006-01-02 15:04:05")),
		"mtime":    starlark.String(mtime.Format("2006-01-02 15:04:05")),
		"ctime":    starlark.String(ctime.Format("2006-01-02 15:04:05")),
	})
}

func Append(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	filePath, err := params.GetString(0)
	if err != nil {
		return starlark.None, err
	}
	content, err := params.GetString(1)
	if err != nil {
		return starlark.None, err
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return starlark.Bool(false), err
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return starlark.Bool(false), nil
	}

	return starlark.Bool(true), nil
}

func Create(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	filePath, err := params.GetString(0)
	if err != nil {
		return starlark.None, err
	}
	content, err := params.GetString(1)
	if err != nil {
		content = ""
	}

	// 如果文件不存在则创建，如果已经存在则会覆盖
	file, err := os.Create(filePath)
	if err != nil {
		return starlark.None, err
	}
	defer file.Close()

	// Write the content to the file
	if _, err := file.WriteString(content); err != nil {
		return starlark.None, err
	}
	return starlark.Bool(true), nil
}

func ReadAll(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	filePath, err := params.GetString(0)
	if err != nil {
		return starlark.None, err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return starlark.None, err
	}
	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return starlark.None, err
	}
	return starlark.String(string(content)), nil
}

func Glob(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	filePath, err := params.GetString(0)
	if err != nil {
		return starlark.None, err
	}

	// 搜索所有以 .txt 结尾的文件
	files, err := filepath.Glob(filePath)
	if err != nil {
		return starlark.None, err
	}
	return util.ConvertToStarlark(files)
}

func Stat(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	filePath, err := params.GetString(0)
	if err != nil {
		return starlark.None, err
	}
	fileStat, err := os.Stat(filePath)
	if err != nil {
		return starlark.None, err
	}
	return convertFileInfo(fileStat), nil
}

func Exist(thread *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	params, err := util.GetParser(args, kwargs)
	if err != nil {
		return starlark.None, err
	}
	filePath, err := params.GetString(0)
	if err != nil {
		return starlark.None, err
	}
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return starlark.Bool(false), err
	}
	return starlark.Bool(true), err
}
