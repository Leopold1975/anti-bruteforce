package ip

import (
	"errors"
	"strconv"
	"strings"
)

func ToInt(ipAddr string) (uint32, error) {
	nums := strings.Split(ipAddr, ".")
	if len(nums) < 4 {
		return 0, errors.New("invalid address")
	}
	var ip uint32
	offset := 24
	for _, n := range nums {
		i, err := strconv.ParseUint(n, 10, 64)
		if err != nil {
			return 0, err
		}
		if i > 255 {
			return 0, errors.New("invalid address")
		}
		ip += uint32(i << offset)
		offset -= 8
	}
	return ip, nil
}

func IntToIP(ip uint32) string {
	var ipAddr string
	for i := 0; i < 4; i++ {
		num := ip % 256
		ip /= 256
		if i == 3 {
			ipAddr = strconv.Itoa(int(num)) + ipAddr
			continue
		}
		ipAddr = "." + strconv.Itoa(int(num)) + ipAddr
	}
	return ipAddr
}

// returns if ip belongs to network.
func BelongsToNetwork(network string, ipAddr string) (bool, error) {
	if len(network) < 7 || len(ipAddr) < 7 {
		return false, errors.New("invalid address")
	}
	s := strings.Split(network, "/")
	if len(s) < 2 {
		s = append(s, "32")
	}
	net, err := ToInt(s[0])
	if err != nil {
		return false, err
	}
	ip, err := ToInt(ipAddr)
	if err != nil {
		return false, err
	}

	mask, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		return false, err
	}

	op := (1<<32 - 1) ^ (1<<(32-mask) - 1)
	return net&uint32(op) == ip&uint32(op), nil
}
