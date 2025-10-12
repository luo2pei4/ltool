package utils

import "strings"

type HostnamectlResult struct {
	IPAddress       string
	Hostname        string
	Architecture    string
	OperationSystem string
	Kernel          string
}

func ParseHostnamectlResult(data []byte) (*HostnamectlResult, error) {
	if len(data) == 0 {
		return nil, nil
	}
	items := strings.Split(string(data), "\n")
	hr := &HostnamectlResult{}
	for _, item := range items {
		tmp := strings.TrimSpace(item)
		switch {
		case strings.HasPrefix(tmp, "Static hostname"):
			hr.Hostname = strings.TrimPrefix(tmp, "Static hostname: ")
		case strings.HasPrefix(tmp, "Operating System"):
			hr.OperationSystem = strings.TrimPrefix(tmp, "Operating System: ")
		case strings.HasPrefix(tmp, "Architecture"):
			hr.Architecture = strings.TrimPrefix(tmp, "Architecture: ")
		case strings.HasPrefix(tmp, "Kernel"):
			hr.Kernel = strings.TrimPrefix(tmp, "Kernel: ")
		default:
		}
	}
	return hr, nil
}
