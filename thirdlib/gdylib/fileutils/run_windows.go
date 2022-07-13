package fileutils

import "os/exec"

//OpenProgramOrFile 启动程序
func OpenProgramOrFile(argv []string) error {

	var startArgvs []string

	startArgvs = append(startArgvs, "/C")
	startArgvs = append(startArgvs, "start")

	startArgvs = append(startArgvs, argv...)

	cmd := exec.Command("cmd.exe", startArgvs...)
	return cmd.Start()
}
