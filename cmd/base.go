package cmd

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

// BindViper 请求参数绑定到全局对象viper上
func BindViper(flags *pflag.FlagSet, names ...string) {
	for _, name := range names {
		err := viper.BindPFlag(name, flags.Lookup(name))
		if err != nil {
			panic(err)
		}
	}
}
