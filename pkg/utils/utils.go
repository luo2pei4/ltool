package utils

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/luo2pei4/ltool/pkg/consts"
	probing "github.com/prometheus-community/pro-bing"
)

// ValidateIPv4 validate ipv4 address
// support xxx.xxx.xxx.xxx-xxx format, "-xxx" mains 'to xxx'
func ValidateIPv4(ip string) error {
	arr := strings.Split(ip, "-")
	if len(arr) > 2 {
		return errors.New("unsupported ip address format")
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

func Ping(ip string) (bool, error) {
	pinger, err := probing.NewPinger(ip)
	if err != nil {
		return false, err
	}
	pinger.Count = 3                 // sends and receives three packets
	pinger.Timeout = time.Second * 3 // timeout
	pinger.SetPrivileged(true)
	err = pinger.Run()
	if err != nil {
		return false, err
	}
	stats := pinger.Statistics()
	if stats.PacketsRecv > 0 {
		return true, nil
	}
	return false, nil
}
