// this package supply some common use tools
package common

import "strconv"

// HumanSize2Bytes parse human readable size to
// machine bytes
func HumanSize2Bytes(size string) int64 {

	weight := size[len(size)-1]
	ret, err := strconv.Atoi(size[:len(size)-1])
	if err != nil {
		panic(ret)
	}
	ret64 := int64(ret)
	if weight == 'B' {
		return ret64
	} else if weight == 'K' {
		return ret64 * (1 << 10)
	} else if weight == 'M' {
		return ret64 * (1 << 20)
	} else if weight == 'G' {
		return ret64 * (1 << 30)
	}

	return ret64
}
