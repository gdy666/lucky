package stringsp

func StrIsInList(str string, strList []string) bool {
	checkMap := make(map[string]uint8)
	for i := range strList {
		checkMap[strList[i]] = 1
	}
	if _, ok := checkMap[str]; ok {
		return true
	}
	return false
}
