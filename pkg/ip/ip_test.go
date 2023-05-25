package ip

import (
	"testing"

	"github.com/stretchr/testify/require"
)

var testData = []string{
	"0.0.0.0",
	"255.255.255.255",
	"127.0.0.1",
	"1.1.1.1",
	"1.2.3.4",
	"127.128.129.1",
	"1.255.2.1",
	"0.0.0.1",
}

func TestIPToInt(t *testing.T) {
	for _, tc := range testData {
		i, err := ToInt(tc)
		require.NoError(t, err)

		addr := IntToIP(i)
		require.Equal(t, tc, addr)
	}
}

type TestCase struct {
	network  string
	ip       string
	expected bool
}

var testCases = []TestCase{
	{
		network:  "0.0.0.0/0",
		ip:       "172.162.32.1",
		expected: true,
	},
	{
		network:  "0.0.0.0/32",
		ip:       "0.0.0.0",
		expected: true,
	},
	{
		network:  "0.0.0.0/32",
		ip:       "0.0.0.1",
		expected: false,
	},
	{
		network:  "172.162.31.1/24",
		ip:       "172.162.31.1",
		expected: true,
	},
	{
		network:  "172.162.31.1/24",
		ip:       "172.162.31.125",
		expected: true,
	},
	{
		network:  "172.162.31.1/24",
		ip:       "172.162.32.1",
		expected: false,
	},
	{
		network:  "172.162.31.1/24",
		ip:       "172.162.32.1",
		expected: false,
	},
	{
		network:  "172.162.31.1/16",
		ip:       "172.162.2.12",
		expected: true,
	},
	{
		network:  "172.162.31.1/8",
		ip:       "172.12.2.12",
		expected: true,
	},
	{
		network:  "172.162.31.1/8",
		ip:       "72.162.2.12",
		expected: false,
	},
	{
		network:  "172.162.31.1",
		ip:       "172.162.31.1",
		expected: true,
	},
	{
		network:  "172.162.31.1",
		ip:       "172.162.31.2",
		expected: false,
	},
	{
		network:  "44.44.44.234/24",
		ip:       "44.44.44.2",
		expected: true,
	},
}

var errorCases = []TestCase{
	{
		network:  "",
		ip:       "72.162.2.12",
		expected: false,
	},
	{
		network:  "172.162.31.1/8",
		ip:       "",
		expected: false,
	},
	{
		network:  "172.31.1",
		ip:       "172.162.31.1",
		expected: false,
	},
	{
		network:  "172.162.31.1",
		ip:       "172.31.1",
		expected: false,
	},
	{
		network:  "172a.162.31.1",
		ip:       "172.162.31.1",
		expected: false,
	},
	{
		network:  "172.162.31.1",
		ip:       "172.1a62.31.1",
		expected: false,
	},
	{
		network:  "172.162.31.1/2a",
		ip:       "172.162.31.1",
		expected: false,
	},
}

func TestIPBelongsToNetwork(t *testing.T) {
	for _, tc := range testCases {
		b, err := BelongsToNetwork(tc.network, tc.ip)
		require.NoError(t, err)
		require.Equal(t, tc.expected, b)
	}
	for _, ec := range errorCases {
		b, err := BelongsToNetwork(ec.network, ec.ip)
		require.NotNil(t, err)
		require.Equal(t, ec.expected, b)
	}
}
