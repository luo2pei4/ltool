package node

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/luo2pei4/ltool/pkg/consts"
)

func validateIP(ip string) error {
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

func addNode(records *[]record, ip, user, password string) {
	arr := strings.Split(ip, "-")
	if len(arr) == 1 {
		*records = append(*records, record{
			ip:       ip,
			user:     user,
			password: password,
			status:   "offline",
		})
		return
	}
	toNodeIP, _ := strconv.Atoi(arr[1])
	tmp := strings.Split(arr[0], ".")
	fromNodeIP, _ := strconv.Atoi(tmp[3])
	if toNodeIP == fromNodeIP {
		*records = append(*records, record{
			ip:       ip,
			user:     user,
			password: password,
			status:   "offline",
		})
		return
	}
	tmpMap := make(map[string]struct{})
	for _, rec := range *records {
		tmpMap[rec.ip] = struct{}{}
	}
	if toNodeIP < fromNodeIP {
		fromNodeIP, toNodeIP = toNodeIP, fromNodeIP
	}
	for ; fromNodeIP <= toNodeIP; fromNodeIP++ {
		tmp[3] = strconv.Itoa(fromNodeIP)
		newIP := strings.Join(tmp, ".")
		if _, ok := tmpMap[newIP]; ok {
			continue
		}
		*records = append(*records, record{
			ip:       strings.Join(tmp, "."),
			user:     user,
			password: password,
			status:   "offline",
		})
	}
}

func makeSelectedStatsMsg(records *[]record) string {
	total := len(*records)
	selected := 0
	for _, rec := range *records {
		if rec.checked {
			selected++
		}
	}
	return fmt.Sprintf("Selected: %d, Total: %d", selected, total)
}
