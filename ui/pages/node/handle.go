package node

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/luo2pei4/ltool/pkg/consts"
	probing "github.com/prometheus-community/pro-bing"
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
	sync.RWMutex
	records     []node
	ipsCh       chan []string
	statusChgCh chan struct{}
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

func (n *nodes) startStatusMonitor() {
	timer := time.NewTimer(time.Second)
	var ipList []string
	for {
		select {
		case <-timer.C:
			n.RLock()
			ipList = make([]string, 0, len(n.records))
			for _, rec := range n.records {
				ipList = append(ipList, rec.ip)
			}
			n.RUnlock()
		case ips := <-n.ipsCh:
			ipList = ips
		}
		n.detectStatus(ipList)
		timer.Reset(time.Second * 60)
	}
}

func (n *nodes) detectStatus(ipList []string) {
	var wg sync.WaitGroup
	resultCh := make(chan string, len(ipList))
	defer close(resultCh)

	for _, ip := range ipList {
		wg.Add(1)
		go func(ip string) {
			defer wg.Done()
			status := "offline"
			if err := pingHost(ip, 3); err != nil {
				status = "online"
			}
			resultCh <- ip + "-" + status
		}(ip)
	}
	wg.Wait()
	cnt := 0
	statusMap := make(map[string]string, len(ipList))
	for res := range resultCh {
		arr := strings.Split(res, "-")
		statusMap[arr[0]] = arr[1]
		cnt++
		if cnt == len(ipList) {
			break
		}
	}
	n.Lock()
	defer n.Unlock()
	changed := false
	for idx, rec := range n.records {
		if status, ok := statusMap[rec.ip]; ok {
			if n.records[idx].status != status {
				fmt.Printf("status changed, ip: %s, new status: %s\n", rec.ip, status)
				n.records[idx].status = status
				changed = true
			}
		}
	}
	if changed {
		n.statusChgCh <- struct{}{}
	}
}

func pingHost(host string, count int) error {
	pinger, err := probing.NewPinger(host)
	if err != nil {
		return err
	}
	// Interval
	pinger.Interval = time.Second
	// time
	pinger.Timeout = time.Second * 3
	// package count
	pinger.Count = count
	return pinger.Run()
}
