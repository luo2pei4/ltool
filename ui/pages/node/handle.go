package node

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/luo2pei4/ltool/pkg/consts"
)

type node struct {
	ip       string
	user     string
	password string
	status   string
	checked  bool
	newRec   bool
	changed  bool
}

type nodes struct {
	records  []node
	ipCh     chan string
	statusCh chan string
}

func (n *nodes) addNode(ip, user, password string) {
	arr := strings.Split(ip, "-")
	if len(arr) == 1 {
		n.records = append(n.records, node{
			ip:       ip,
			user:     user,
			password: password,
			status:   "offline",
			newRec:   true,
		})
		return
	}
	toNodeIP, _ := strconv.Atoi(arr[1])
	tmp := strings.Split(arr[0], ".")
	fromNodeIP, _ := strconv.Atoi(tmp[3])
	if toNodeIP == fromNodeIP {
		n.records = append(n.records, node{
			ip:       ip,
			user:     user,
			password: password,
			status:   "offline",
			newRec:   true,
		})
		return
	}
	tmpMap := make(map[string]struct{})
	for _, rec := range n.records {
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
		n.records = append(n.records, node{
			ip:       strings.Join(tmp, "."),
			user:     user,
			password: password,
			status:   "offline",
			newRec:   true,
		})
	}
}

func (n *nodes) makeSelectedStatsMsg() string {
	total := len(n.records)
	selected := 0
	newRecs := 0
	changed := 0
	for _, rec := range n.records {
		if rec.checked {
			selected++
		}
		if rec.newRec {
			newRecs++
		} else if rec.changed {
			changed++
		}
	}
	return fmt.Sprintf("total: %d, new: %d, changed: %d, selected: %d", total, newRecs, changed, selected)
}

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
