package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/luo2pei4/ltool/pkg/consts"
)

func ValidateIP(ip string) error {
	// check ip address format
	// support xxx.xxx.xxx.xxx-x format, "-x" mains 'to x'
	arr := strings.Split(ip, "-")
	if len(arr) > 2 {
		return errors.New("invalid ip address format")
	}
	if len(arr) == 1 || len(arr) == 2 {
		matched, err := regexp.MatchString(consts.IPv4Pattern, arr[0])
		if err != nil {
			return fmt.Errorf("check ip address failed, %v", err)
		}
		if !matched {
			return fmt.Errorf("invalid ip address")
		}
	}
	if len(arr) == 2 {
		toNodeIP, err := strconv.Atoi(arr[1])
		if err != nil {
			return err
		}
		if toNodeIP > 255 {
			return errors.New("invalid ip address range")
		}
	}
	return nil
}
