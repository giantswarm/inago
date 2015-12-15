package infraconfigparser

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type StringSlice []string

func (ss *StringSlice) String() string {
	return fmt.Sprintf("%s", *ss)
}

func (ss *StringSlice) Set(val string) error {
	*ss = append(*ss, val)
	return nil
}

func (ss *StringSlice) Type() string {
	return "StringSlice"
}

func (ss *StringSlice) Viper() *viper.Viper {
	v := viper.New()
	for _, val := range *ss {
		splitted := strings.Split(val, ":")
		if len(splitted) == 2 {
			v.Set(splitted[0], splitted[1])
		}
	}
	return v
}
