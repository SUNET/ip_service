package rpsl

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseRouteObjects(t *testing.T) {
	tts := []struct {
		name         string
		filePath     string
		wantRoutes   int
		wantRoute6   int
		wantNetworks []string
		wantOrigins  map[string]string // network -> origin
	}{
		{
			name:       "mix.golden - mixed RPSL objects",
			filePath:   "./testdata/mix.golden",
			wantRoutes: 6,
			wantRoute6: 1,
			wantNetworks: []string{
				"193.254.30.0/24",
				"212.166.64.0/19",
				"195.2.0.0/19",
				"212.44.224.0/19",
				"212.80.176.0/29",
				"2001:1578:100::/40",
			},
			wantOrigins: map[string]string{
				"193.254.30.0/24":    "AS12726",
				"195.2.0.0/19":       "AS1273",
				"2001:1578:100::/40": "AS29317",
			},
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			ctx := t.Context()
			client, err := New(ctx)
			require.NoError(t, err)

			// Reset package-level state that leaks between Parse calls
			interCount = 0

			err = client.Parse(ctx, tt.filePath)
			require.NoError(t, err)

			for _, network := range tt.wantNetworks {
				_, ok := client.RouterClass[network]
				assert.True(t, ok, "expected network %s in RouterClass", network)
			}

			for network, origin := range tt.wantOrigins {
				asn, ok := client.RouterClass[network]
				assert.True(t, ok, "expected network %s in RouterClass", network)
				if ok {
					_, hasOrigin := asn[origin]
					assert.True(t, hasOrigin, "expected origin %s for network %s", origin, network)
				}
			}
		})
	}
}

func TestParseMultiOriginRoute(t *testing.T) {
	// 212.166.64.0/19 appears twice with different origins in mix.golden
	ctx := t.Context()
	client, err := New(ctx)
	require.NoError(t, err)

	err = client.Parse(ctx, "./testdata/mix.golden")
	require.NoError(t, err)

	asn, ok := client.RouterClass["212.166.64.0/19"]
	require.True(t, ok)
	assert.Len(t, asn, 2)
	assert.Contains(t, asn, "AS12321")
	assert.Contains(t, asn, "AS12541")
}

func TestParseNonExistentFile(t *testing.T) {
	ctx := t.Context()
	client, err := New(ctx)
	require.NoError(t, err)

	err = client.Parse(ctx, "./testdata/nonexistent.txt")
	assert.Error(t, err)
}

func TestParseEOFMarker(t *testing.T) {
	content := `route:          10.0.0.0/8
origin:         AS64512

route:          172.16.0.0/12
origin:         AS64513

# EOF
route:          192.168.0.0/16
origin:         AS64514

`
	tmpFile := filepath.Join(t.TempDir(), "eof_test.txt")
	err := os.WriteFile(tmpFile, []byte(content), 0600)
	require.NoError(t, err)

	ctx := t.Context()
	client, err := New(ctx)
	require.NoError(t, err)

	err = client.Parse(ctx, tmpFile)
	require.NoError(t, err)

	assert.Len(t, client.RouterClass, 2)
	_, ok := client.RouterClass["192.168.0.0/16"]
	assert.False(t, ok, "should not parse objects after # EOF")
}

func TestParseCommentSkipping(t *testing.T) {
	content := `# This is a comment
# Another comment
route:          10.0.0.0/8
descr:          Example
origin:         AS64512
# inline comment between attributes won't break parsing

`
	tmpFile := filepath.Join(t.TempDir(), "comment_test.txt")
	err := os.WriteFile(tmpFile, []byte(content), 0600)
	require.NoError(t, err)

	ctx := t.Context()
	client, err := New(ctx)
	require.NoError(t, err)

	err = client.Parse(ctx, tmpFile)
	require.NoError(t, err)

	assert.Len(t, client.RouterClass, 1)
	_, ok := client.RouterClass["10.0.0.0/8"]
	assert.True(t, ok)
}

func TestParseIPv6Route(t *testing.T) {
	content := `route6:         2001:db8::/32
descr:          Test IPv6 network
origin:         AS64512
mnt-by:         TEST-MNT
source:         TEST

route6:         2001:67c:2564::/48
descr:          SUNET network
origin:         AS1653
source:         RIPE

# EOF
`
	tmpFile := filepath.Join(t.TempDir(), "ipv6_test.txt")
	err := os.WriteFile(tmpFile, []byte(content), 0600)
	require.NoError(t, err)

	ctx := t.Context()
	client, err := New(ctx)
	require.NoError(t, err)

	err = client.Parse(ctx, tmpFile)
	require.NoError(t, err)

	assert.Len(t, client.RouterClass, 2)

	asn, ok := client.RouterClass["2001:db8::/32"]
	require.True(t, ok)
	assert.Contains(t, asn, "AS64512")

	asn, ok = client.RouterClass["2001:67c:2564::/48"]
	require.True(t, ok)
	assert.Contains(t, asn, "AS1653")
}

func TestParseContinuationLines(t *testing.T) {
	content := `route:          10.0.0.0/8
descr:          Example network
                with continuation
origin:         AS64512
source:         TEST

# EOF
`
	tmpFile := filepath.Join(t.TempDir(), "continuation_test.txt")
	err := os.WriteFile(tmpFile, []byte(content), 0600)
	require.NoError(t, err)

	ctx := t.Context()
	client, err := New(ctx)
	require.NoError(t, err)

	// Reset package-level state (interCount leaks between Parse calls)
	interCount = 0

	err = client.Parse(ctx, tmpFile)
	require.NoError(t, err)

	require.Len(t, client.RouterClass, 1)
	asn, ok := client.RouterClass["10.0.0.0/8"]
	require.True(t, ok)
	obj := asn["AS64512"]
	require.NotNil(t, obj)
	assert.Equal(t, "10.0.0.0/8", obj.Network)
	assert.Equal(t, "AS64512", obj.Origin)
}

func TestObjectAdd(t *testing.T) {
	tts := []struct {
		name    string
		key     string
		value   string
		check   func(t *testing.T, obj *Object)
		wantErr bool
	}{
		{
			name:  "route sets Network",
			key:   Route,
			value: "192.168.0.0/16",
			check: func(t *testing.T, obj *Object) {
				assert.Equal(t, "192.168.0.0/16", obj.Network)
			},
		},
		{
			name:  "route6 sets Network",
			key:   Route6,
			value: "2001:db8::/32",
			check: func(t *testing.T, obj *Object) {
				assert.Equal(t, "2001:db8::/32", obj.Network)
			},
		},
		{
			name:    "empty route returns error",
			key:     Route,
			value:   "",
			wantErr: true,
		},
		{
			name:    "empty route6 returns error",
			key:     Route6,
			value:   "",
			wantErr: true,
		},
		{
			name:  "origin sets Origin",
			key:   Origin,
			value: "AS64512",
			check: func(t *testing.T, obj *Object) {
				assert.Equal(t, "AS64512", obj.Origin)
			},
		},
		{
			name:  "descr is ignored",
			key:   Descr,
			value: "first line",
			check: func(t *testing.T, obj *Object) {
				// descr is no longer stored
			},
		},
		{
			name:  "source is ignored",
			key:   Source,
			value: "RIPE",
			check: func(t *testing.T, obj *Object) {
				// source is no longer stored
			},
		},
		{
			name:  "mnt-by is ignored",
			key:   MNTBy,
			value: "SUNET-MNT",
			check: func(t *testing.T, obj *Object) {
				// mnt-by is no longer stored
			},
		},
		{
			name:  "country appends",
			key:   Country,
			value: "SE",
			check: func(t *testing.T, obj *Object) {
				assert.Contains(t, obj.Country, "SE")
			},
		},
		{
			name:  "last-modified sets string",
			key:   LastModified,
			value: "2024-01-01T00:00:00Z",
			check: func(t *testing.T, obj *Object) {
				assert.Equal(t, "2024-01-01T00:00:00Z", obj.LastModified)
			},
		},
		{
			name:  "empty key is no-op",
			key:   "",
			value: "anything",
			check: func(t *testing.T, obj *Object) {
				// Should not panic or change anything
			},
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			obj := &Object{}
			err := obj.Add(tt.key, tt.value)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			if tt.check != nil {
				tt.check(t, obj)
			}
		})
	}
}

func TestObjectFindNetwork(t *testing.T) {
	tts := []struct {
		name    string
		network string
		ip      string
		want    bool
		wantErr bool
	}{
		{
			name:    "IPv4 - inside /24",
			network: "192.168.1.0/24",
			ip:      "192.168.1.100",
			want:    true,
		},
		{
			name:    "IPv4 - outside /24",
			network: "192.168.1.0/24",
			ip:      "192.168.2.1",
			want:    false,
		},
		{
			name:    "IPv4 - first address",
			network: "10.0.0.0/8",
			ip:      "10.0.0.0",
			want:    true,
		},
		{
			name:    "IPv4 - last address",
			network: "10.0.0.0/8",
			ip:      "10.255.255.255",
			want:    true,
		},
		{
			name:    "IPv6 - inside /48",
			network: "2001:db8:1::/48",
			ip:      "2001:db8:1::1",
			want:    true,
		},
		{
			name:    "IPv6 - outside /48",
			network: "2001:db8:1::/48",
			ip:      "2001:db8:2::1",
			want:    false,
		},
		{
			name:    "invalid network",
			network: "not-a-network",
			ip:      "192.168.1.1",
			wantErr: true,
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			obj := &Object{Network: tt.network}
			got, err := obj.FindNetwork(t.Context(), tt.ip)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClientGetKey(t *testing.T) {
	tts := []struct {
		name       string
		line       string
		currentKey string
		want       string
	}{
		{
			name: "simple key",
			line: "route:          192.168.0.0/16",
			want: "route",
		},
		{
			name: "route6 key",
			line: "route6:         2001:db8::/32",
			want: "route6",
		},
		{
			name: "origin key",
			line: "origin:         AS64512",
			want: "origin",
		},
		{
			name:       "continuation line returns current key",
			line:       "                continued value",
			currentKey: "descr",
			want:       "descr",
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(t.Context())
			require.NoError(t, err)

			if tt.currentKey != "" {
				client.currentKey = &tt.currentKey
			}

			got := client.getKey(tt.line)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestClientGetValue(t *testing.T) {
	tts := []struct {
		name string
		line string
		key  string
		want string
	}{
		{
			name: "route value",
			line: "route:          192.168.0.0/16",
			key:  "route",
			want: "192.168.0.0/16",
		},
		{
			name: "route6 value",
			line: "route6:         2001:db8::/32",
			key:  "route6",
			want: "2001:db8::/32",
		},
		{
			name: "origin value",
			line: "origin:         AS64512",
			key:  "origin",
			want: "AS64512",
		},
		{
			name: "descr with spaces",
			line: "descr:          Example network description",
			key:  "descr",
			want: "Example network description",
		},
		{
			name: "continuation line value",
			line: "                continued text here",
			key:  "descr",
			want: "continued text here",
		},
	}

	for _, tt := range tts {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(t.Context())
			require.NoError(t, err)

			got := client.getValue(tt.line, tt.key)
			assert.Equal(t, tt.want, got)
		})
	}
}
