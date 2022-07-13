package fileutils

import (
	"os/exec"
)

//OpenProgramOrFile 启动程序
func OpenProgramOrFile(argv []string) error {

	var startArgvs []string

	for i := range argv {
		if i == 0 {
			continue
		}
		startArgvs = append(startArgvs, argv[i])
	}
	//startArgvs = append(startArgvs, "-c")
	//startArgvs = append(startArgvs, argv...)

	//fmt.Printf("fuck...%v \n", startArgvs)

	//cmd := exec.Command("/bin/bash", startArgvs...)
	cmd := exec.Command(argv[0], startArgvs...)
	return cmd.Start()
}
