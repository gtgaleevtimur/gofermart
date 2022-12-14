package loon

import (
	"math"
	"strconv"
	"strings"
)

func IsValid(number string) bool {
	digits := strings.Split(strings.ReplaceAll(number, " ", ""), "")
	lengthOfString := len(digits)

	if lengthOfString < 2 {
		return false
	}

	sum := 0
	flag := false

	for i := lengthOfString - 1; i > -1; i-- {
		digit, _ := strconv.Atoi(digits[i])

		if flag {
			digit *= 2

			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		flag = !flag
	}

	return math.Mod(float64(sum), 10) == 0
}
