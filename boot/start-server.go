package boot

import (
	"fmt"
) // with go modules enabled (GO111MODULE=on or outside GOPATH)

func StartServer(cfg *GitServerConfigs) {
	fmt.Println("will start git server:", cfg)
}
