package rule

import (
	"testing"
)

func Test_createSubRuleListFromConfigure_1(t *testing.T) {
	argsA := "53to192.168.31.1"

	goportsList, _, _, _, _, _, _, e := createSubRuleListFromConfigure(argsA)
	if e != nil || len(goportsList) != 2 {
		t.Errorf("createSubRuleListFromConfigure [%s] testA error %v", argsA, goportsList)
		return
	}

	if goportsList[0].ProxyType != "tcp" || len(goportsList[0].ListenPorts) != 1 || goportsList[0].TargetHost != "192.168.31.1" || len(goportsList[0].TargetPorts) != 1 || len(goportsList[0].ListenPorts) != len(goportsList[0].TargetPorts) {
		t.Errorf("createSubRuleListFromConfigure [%s] testB error %v", argsA, goportsList)
		return
	}

	if goportsList[0].ListenPorts[0] != 53 || goportsList[0].TargetPorts[0] != 53 {
		t.Errorf("createSubRuleListFromConfigure [%s] testC error %v", argsA, goportsList)
		return
	}

	if goportsList[1].ProxyType != "udp" || len(goportsList[0].ListenPorts) != 1 || goportsList[0].TargetHost != "192.168.31.1" || len(goportsList[0].TargetPorts) != 1 || len(goportsList[0].ListenPorts) != len(goportsList[0].TargetPorts) {
		t.Errorf("createSubRuleListFromConfigure [%s] testD error %v", argsA, goportsList)
		return
	}

	if goportsList[1].ListenPorts[0] != 53 || goportsList[1].TargetPorts[0] != 53 {
		t.Errorf("createSubRuleListFromConfigure [%s] testE error %v", argsA, goportsList)
		return
	}

}

func Test_createSubRuleListFromConfigure_2(t *testing.T) {
	argsA := "tcp@53to192.168.31.1"

	goportsList, _, _, _, _, _, _, e := createSubRuleListFromConfigure(argsA)
	if e != nil || len(goportsList) != 1 {
		t.Errorf("createSubRuleListFromConfigure [%s] testA error %v", argsA, goportsList)
		return
	}

	if goportsList[0].ProxyType != "tcp" || len(goportsList[0].ListenPorts) != 1 || goportsList[0].TargetHost != "192.168.31.1" || len(goportsList[0].TargetPorts) != 1 || len(goportsList[0].ListenPorts) != len(goportsList[0].TargetPorts) {
		t.Errorf("createSubRuleListFromConfigure [%s] testB error %v", argsA, goportsList)
		return
	}

	if goportsList[0].ListenPorts[0] != 53 || goportsList[0].TargetPorts[0] != 53 {
		t.Errorf("createSubRuleListFromConfigure [%s] testC error %v", argsA, goportsList)
		return
	}

}

func Test_createSubRuleListFromConfigure_3(t *testing.T) {
	argsA := "udp4@53to192.168.31.1"

	goportsList, _, _, _, _, _, _, e := createSubRuleListFromConfigure(argsA)
	if e != nil || len(goportsList) != 1 {
		t.Errorf("createSubRuleListFromConfigure [%s] testA error %v", argsA, goportsList)
		return
	}

	if goportsList[0].ProxyType != "udp4" || len(goportsList[0].ListenPorts) != 1 || goportsList[0].TargetHost != "192.168.31.1" || len(goportsList[0].TargetPorts) != 1 || len(goportsList[0].ListenPorts) != len(goportsList[0].TargetPorts) {
		t.Errorf("createSubRuleListFromConfigure [%s] testB error %v", argsA, goportsList)
		return
	}

	if goportsList[0].ListenPorts[0] != 53 || goportsList[0].TargetPorts[0] != 53 {
		t.Errorf("createSubRuleListFromConfigure [%s] testC error %v", argsA, goportsList)
		return
	}

}

func Test_createSubRuleListFromConfigure_4_0(t *testing.T) {
	argsList := []string{"tcp@20000-20021to192.168.31.1", "tcp@192.168.31.1:20000-20021", "tcp@20000-20021to192.168.31.1:20000-20021"}

	for _, argsA := range argsList {
		goportsList, _, _, _, _, _, _, e := createSubRuleListFromConfigure(argsA)
		if e != nil || len(goportsList) != 1 {
			t.Errorf("createSubRuleListFromConfigure [%s] testA error %v", argsA, goportsList)
			return
		}

		if goportsList[0].ProxyType != "tcp" || len(goportsList[0].ListenPorts) != 22 || goportsList[0].TargetHost != "192.168.31.1" || len(goportsList[0].TargetPorts) != 22 || len(goportsList[0].ListenPorts) != len(goportsList[0].TargetPorts) {
			t.Errorf("createSubRuleListFromConfigure [%s] testB error %v", argsA, goportsList)
			return
		}

		if goportsList[0].ListenPorts[0] != 20000 || goportsList[0].TargetPorts[0] != 20000 {
			t.Errorf("createSubRuleListFromConfigure [%s] testC error %v", argsA, goportsList)
			return
		}

		if goportsList[0].ListenPorts[21] != 20021 || goportsList[0].TargetPorts[21] != 20021 {
			t.Errorf("createSubRuleListFromConfigure [%s] testD error %v", argsA, goportsList)
			return
		}
	}

}

func Test_createSubRuleListFromConfigure_4_1(t *testing.T) {
	argsA := "tcp@30000-30021to192.168.31.1:20000-20021"

	goportsList, _, _, _, _, _, _, e := createSubRuleListFromConfigure(argsA)
	if e != nil || len(goportsList) != 1 {
		t.Errorf("createSubRuleListFromConfigure [%s] testA error %v", argsA, goportsList)
		return
	}

	if goportsList[0].ProxyType != "tcp" || len(goportsList[0].ListenPorts) != 22 || goportsList[0].TargetHost != "192.168.31.1" || len(goportsList[0].TargetPorts) != 22 || len(goportsList[0].ListenPorts) != len(goportsList[0].TargetPorts) {
		t.Errorf("createSubRuleListFromConfigure [%s] testB error %v", argsA, goportsList)
		return
	}

	if goportsList[0].ListenPorts[0] != 30000 || goportsList[0].TargetPorts[0] != 20000 {
		t.Errorf("createSubRuleListFromConfigure [%s] testC error %v", argsA, goportsList)
		return
	}

	if goportsList[0].ListenPorts[21] != 30021 || goportsList[0].TargetPorts[21] != 20021 {
		t.Errorf("createSubRuleListFromConfigure [%s] testD error %v", argsA, goportsList)
		return
	}

}

func Test_createSubRuleListFromConfigure_5_0(t *testing.T) {
	//argsA :=
	args := []string{"tcp@80,443,20000-20021to192.168.31.1:80,443,20000-20021",
		"tcp@192.168.31.1:80,443,20000-20021",
		"tcp@80,443,20000-20021to192.168.31.1"}

	for _, argsA := range args {
		goportsList, _, _, _, _, _, _, e := createSubRuleListFromConfigure(argsA)
		if e != nil || len(goportsList) != 1 {
			t.Errorf("createSubRuleListFromConfigure [%s] testA error %v", argsA, goportsList)
			return
		}

		if goportsList[0].ProxyType != "tcp" || len(goportsList[0].ListenPorts) != 24 || goportsList[0].TargetHost != "192.168.31.1" || len(goportsList[0].TargetPorts) != 24 || len(goportsList[0].ListenPorts) != len(goportsList[0].TargetPorts) {
			t.Errorf("createSubRuleListFromConfigure [%s] testB error %v", argsA, goportsList)
			return
		}

		if goportsList[0].ListenPorts[0] != 80 || goportsList[0].TargetPorts[0] != 80 {
			t.Errorf("createSubRuleListFromConfigure [%s] testC error %v", argsA, goportsList)
			return
		}

		if goportsList[0].ListenPorts[23] != 20021 || goportsList[0].TargetPorts[23] != 20021 {
			t.Errorf("createSubRuleListFromConfigure [%s] testD error %v", argsA, goportsList)
			return
		}
	}
}

func Test_createSubRuleListFromConfigure_5_1(t *testing.T) {
	//argsA :=
	args := []string{"tcp@80,443,30000-30021to192.168.31.1:81,443,20000-20021"}

	for _, argsA := range args {
		goportsList, _, _, _, _, _, _, e := createSubRuleListFromConfigure(argsA)
		if e != nil || len(goportsList) != 1 {
			t.Errorf("createSubRuleListFromConfigure [%s] testA error %v", argsA, goportsList)
			return
		}

		if goportsList[0].ProxyType != "tcp" || len(goportsList[0].ListenPorts) != 24 || goportsList[0].TargetHost != "192.168.31.1" || len(goportsList[0].TargetPorts) != 24 || len(goportsList[0].ListenPorts) != len(goportsList[0].TargetPorts) {
			t.Errorf("createSubRuleListFromConfigure [%s] testB error %v", argsA, goportsList)
			return
		}

		if goportsList[0].ListenPorts[0] != 80 || goportsList[0].TargetPorts[0] != 81 {
			t.Errorf("createSubRuleListFromConfigure [%s] testC error %v", argsA, goportsList)
			return
		}

		if goportsList[0].ListenPorts[23] != 30021 || goportsList[0].TargetPorts[23] != 20021 {
			t.Errorf("createSubRuleListFromConfigure [%s] testD error %v", argsA, goportsList)
			return
		}
	}

}

func Test_createSubRuleListFromConfigure_6_0(t *testing.T) {
	//argsA :=
	args := []string{"tcp6@80,443to192.168.31.1:80,443",
		"tcp6@80,443to192.168.31.1",
		"tcp6@192.168.31.1:80,443"}

	for _, argsA := range args {
		goportsList, _, _, _, _, _, _, e := createSubRuleListFromConfigure(argsA)
		if e != nil || len(goportsList) != 1 {
			t.Errorf("createSubRuleListFromConfigure [%s] testA error %v", argsA, goportsList)
			return
		}

		if goportsList[0].ProxyType != "tcp6" || len(goportsList[0].ListenPorts) != 2 || goportsList[0].TargetHost != "192.168.31.1" || len(goportsList[0].TargetPorts) != 2 || len(goportsList[0].ListenPorts) != len(goportsList[0].TargetPorts) {
			t.Errorf("createSubRuleListFromConfigure [%s] testB error %v", argsA, goportsList)
			return
		}

		if goportsList[0].ListenPorts[0] != 80 || goportsList[0].TargetPorts[0] != 80 {
			t.Errorf("createSubRuleListFromConfigure [%s] testC error %v", argsA, goportsList)
			return
		}

		if goportsList[0].ListenPorts[1] != 443 || goportsList[0].TargetPorts[1] != 443 {
			t.Errorf("createSubRuleListFromConfigure [%s] testD error %v", argsA, goportsList)
			return
		}
	}

}

func Test_createSubRuleListFromConfigure_6_1(t *testing.T) {
	//argsA :=
	args := []string{"tcp@80,443to192.168.31.1:81,443"}

	for _, argsA := range args {
		goportsList, _, _, _, _, _, _, e := createSubRuleListFromConfigure(argsA)
		if e != nil || len(goportsList) != 1 {
			t.Errorf("createSubRuleListFromConfigure [%s] testA error %v", argsA, goportsList)
			return
		}

		if goportsList[0].ProxyType != "tcp" || len(goportsList[0].ListenPorts) != 2 || goportsList[0].TargetHost != "192.168.31.1" || len(goportsList[0].TargetPorts) != 2 || len(goportsList[0].ListenPorts) != len(goportsList[0].TargetPorts) {
			t.Errorf("createSubRuleListFromConfigure [%s] testB error %v", argsA, goportsList)
			return
		}

		if goportsList[0].ListenPorts[0] != 80 || goportsList[0].TargetPorts[0] != 81 {
			t.Errorf("createSubRuleListFromConfigure [%s] testC error %v", argsA, goportsList)
			return
		}

		if goportsList[0].ListenPorts[1] != 443 || goportsList[0].TargetPorts[1] != 443 {
			t.Errorf("createSubRuleListFromConfigure [%s] testD error %v", argsA, goportsList)
			return
		}
	}

}
