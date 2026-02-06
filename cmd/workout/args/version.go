package args

import (
	"fmt"
	"os"
)

type Version struct {
	version string
}

func (v *Version) SetVersion(cli *Cli) string {
	fmt.Println(v.version)

	os.Exit(0)
	return ""
}
