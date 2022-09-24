package stringsp

import "fmt"

const (
	KB = uint(1024)
	MB = uint64(1024 * 1024)
	GB = uint64(1024 * 1024 * 1024)
	TB = uint64(1024 * 1024 * 1024 * 1024)
	EB = uint64(1024 * 1024 * 1024 * 1024 * 1024)
)

func BinaryUnitToStr(binaryUnit uint64) (size string) {
	switch {
	case binaryUnit < 1024:
		return fmt.Sprintf("%.2fB", float64(binaryUnit)/float64(1))
	case binaryUnit < MB: //1024*1024
		return fmt.Sprintf("%.2fKB", float64(binaryUnit)/float64(KB))
	case binaryUnit < GB: //1024 * 1024 * 1024
		return fmt.Sprintf("%.2fMB", float64(binaryUnit)/float64(MB))
	case binaryUnit < TB: //1024 * 1024 * 1024 * 1024
		return fmt.Sprintf("%.2fGB", float64(binaryUnit)/float64(GB))
	case binaryUnit < EB: //1024 * 1024 * 1024 * 1024 * 1024
		return fmt.Sprintf("%.2fTB", float64(binaryUnit)/float64(TB))
	default:
		return fmt.Sprintf("%.2fEB", float64(binaryUnit)/float64(EB))
	}
}
