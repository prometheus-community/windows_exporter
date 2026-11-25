// SPDX-License-Identifier: MIT
//
// Copyright The Prometheus Authors

//go:build windows

// Command portprobe runs bounded Winsock experiments on how Windows hands out
// ephemeral ports to IPv4/IPv6 and TCP/UDP sockets, and reads the undocumented
// NSI port pool objects. It is a research tool, not part of the exporter.
//
// Modes:
//
//	portprobe nsi              dump NSI objects (read only)
//	portprobe pairs            explicit-bind conflict matrix (no config change)
//	portprobe collide          port 0 collision statistics (no config change)
//	portprobe exhaust          exhaustion experiments; needs a small range that
//	                           the caller set before (see -start/-count)
package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
)

// Socket options not in x/sys/windows.
const (
	ipv6V6Only          = 27
	soExclusiveAddrUse  = ^windows.SO_REUSEADDR // -5
	soReuseUnicastPort  = 0x3007
	soPortScalability   = 0x3006
	maxSocketsPerSeries = 3000
)

type family int

const (
	v4 family = iota
	v6only
	dual
)

func (f family) String() string { return [...]string{"v4", "v6only", "dual"}[f] }

type spec struct {
	fam      family
	proto    string // tcp or udp
	addr     string // wildcard, loopback, loopback2
	excl     bool
	reuse    bool
	unicast  bool // SO_REUSE_UNICASTPORT
	scalable bool // SO_PORT_SCALABILITY
	listen   bool
}

func (s spec) String() string {
	o := fmt.Sprintf("%s/%s/%s", s.proto, s.fam, s.addr)
	if s.excl {
		o += "+EXCL"
	}

	if s.reuse {
		o += "+REUSE"
	}

	if s.unicast {
		o += "+UNICAST"
	}

	if s.scalable {
		o += "+SCALABLE"
	}

	if s.listen {
		o += "+LISTEN"
	}

	return o
}

func (s spec) sockaddr(port int) windows.Sockaddr {
	if s.fam == v4 {
		sa := &windows.SockaddrInet4{Port: port}

		switch s.addr {
		case "loopback":
			sa.Addr = [4]byte{127, 0, 0, 1}
		case "loopback2":
			sa.Addr = [4]byte{127, 0, 0, 2}
		}

		return sa
	}

	sa := &windows.SockaddrInet6{Port: port}
	if s.addr == "loopback" {
		sa.Addr[15] = 1
	}

	if s.addr == "loopback2" && s.fam == dual {
		// ::ffff:127.0.0.2
		copy(sa.Addr[10:], []byte{0xff, 0xff, 127, 0, 0, 2})
	}

	return sa
}

func (s spec) open() (windows.Handle, error) {
	af := windows.AF_INET
	if s.fam != v4 {
		af = windows.AF_INET6
	}

	typ, proto := windows.SOCK_STREAM, windows.IPPROTO_TCP
	if s.proto == "udp" {
		typ, proto = windows.SOCK_DGRAM, windows.IPPROTO_UDP
	}

	h, err := windows.Socket(af, typ, proto)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("socket: %w", err)
	}

	set := func(level, opt, v int, name string) error {
		if err := windows.SetsockoptInt(h, level, opt, v); err != nil {
			windows.Closesocket(h)

			return fmt.Errorf("setsockopt %s: %w", name, err)
		}

		return nil
	}

	if s.fam != v4 {
		v := 1
		if s.fam == dual {
			v = 0
		}

		if err := set(windows.IPPROTO_IPV6, ipv6V6Only, v, "IPV6_V6ONLY"); err != nil {
			return windows.InvalidHandle, err
		}
	}

	if s.excl {
		if err := set(windows.SOL_SOCKET, soExclusiveAddrUse, 1, "SO_EXCLUSIVEADDRUSE"); err != nil {
			return windows.InvalidHandle, err
		}
	}

	if s.reuse {
		if err := set(windows.SOL_SOCKET, windows.SO_REUSEADDR, 1, "SO_REUSEADDR"); err != nil {
			return windows.InvalidHandle, err
		}
	}

	if s.unicast {
		if err := set(windows.SOL_SOCKET, soReuseUnicastPort, 1, "SO_REUSE_UNICASTPORT"); err != nil {
			return windows.InvalidHandle, err
		}
	}

	if s.scalable {
		if err := set(windows.SOL_SOCKET, soPortScalability, 1, "SO_PORT_SCALABILITY"); err != nil {
			return windows.InvalidHandle, err
		}
	}

	return h, nil
}

func localPort(h windows.Handle) int {
	sa, err := windows.Getsockname(h)
	if err != nil {
		return -1
	}

	switch a := sa.(type) {
	case *windows.SockaddrInet4:
		return a.Port
	case *windows.SockaddrInet6:
		return a.Port
	}

	return -1
}

// bindPort opens a socket for s and binds it to port (0 = ephemeral).
func bindPort(s spec, port int) (windows.Handle, int, error) {
	h, err := s.open()
	if err != nil {
		return h, 0, err
	}

	if err := windows.Bind(h, s.sockaddr(port)); err != nil {
		windows.Closesocket(h)

		return windows.InvalidHandle, 0, fmt.Errorf("bind: %w", errName(err))
	}

	if s.listen && s.proto == "tcp" {
		if err := windows.Listen(h, 16); err != nil {
			windows.Closesocket(h)

			return windows.InvalidHandle, 0, fmt.Errorf("listen: %w", errName(err))
		}
	}

	return h, localPort(h), nil
}

type namedErr struct {
	err  error
	name string
}

func (e namedErr) Error() string { return fmt.Sprintf("%s(%d)", e.name, e.err) }
func (e namedErr) Unwrap() error { return e.err }

func errName(err error) error {
	var en windows.Errno
	if !errors.As(err, &en) {
		return err
	}

	names := map[windows.Errno]string{
		10048: "WSAEADDRINUSE", 10049: "WSAEADDRNOTAVAIL", 10055: "WSAENOBUFS",
		10013: "WSAEACCES", 10022: "WSAEINVAL", 10061: "WSAECONNREFUSED",
		10060: "WSAETIMEDOUT", 10047: "WSAEAFNOSUPPORT",
	}
	if n, ok := names[en]; ok {
		return namedErr{err: en, name: n}
	}

	return namedErr{err: en, name: "WSA"}
}

func closeAll(hs []windows.Handle) {
	for _, h := range hs {
		windows.Closesocket(h)
	}
}

// ---------------------------------------------------------------- pairs

func runPairs() {
	tcp := func(f family, addr string) spec { return spec{fam: f, proto: "tcp", addr: addr} }
	udp := func(f family, addr string) spec { return spec{fam: f, proto: "udp", addr: addr} }

	type pair struct{ a, b spec }

	base := []pair{
		{tcp(v4, "wildcard"), tcp(v4, "wildcard")},
		{tcp(v4, "wildcard"), tcp(v6only, "wildcard")},
		{tcp(v4, "wildcard"), tcp(dual, "wildcard")},
		{tcp(v6only, "wildcard"), tcp(v4, "wildcard")},
		{tcp(v6only, "wildcard"), tcp(v6only, "wildcard")},
		{tcp(dual, "wildcard"), tcp(v4, "wildcard")},
		{tcp(v4, "loopback"), tcp(v4, "loopback2")},
		{tcp(v4, "loopback"), tcp(v4, "wildcard")},
		{tcp(v4, "wildcard"), tcp(v4, "loopback")},
		{tcp(v4, "loopback"), tcp(v6only, "loopback")},
		{tcp(v6only, "loopback"), tcp(v4, "loopback")},
		{udp(v4, "wildcard"), udp(v4, "wildcard")},
		{udp(v4, "wildcard"), udp(v6only, "wildcard")},
		{udp(v4, "wildcard"), udp(dual, "wildcard")},
		{udp(v6only, "wildcard"), udp(v4, "wildcard")},
		{udp(dual, "wildcard"), udp(v4, "wildcard")},
		{tcp(v4, "wildcard"), udp(v4, "wildcard")},
		{udp(v4, "wildcard"), tcp(v4, "wildcard")},
		{tcp(v6only, "wildcard"), udp(v6only, "wildcard")},
	}

	fmt.Println("## pairs: first binds port 0, second binds the same port explicitly")
	fmt.Println("first\tsecond\tport\tresult")

	for _, p := range base {
		for _, firstMod := range []string{"", "EXCL", "LISTEN"} {
			for _, secondMod := range []string{"", "REUSE", "EXCL"} {
				a, b := p.a, p.b
				if firstMod == "LISTEN" && a.proto != "tcp" {
					continue
				}

				a.excl = firstMod == "EXCL"
				a.listen = firstMod == "LISTEN"
				b.reuse = secondMod == "REUSE"
				b.excl = secondMod == "EXCL"

				h1, port, err := bindPort(a, 0)
				if err != nil {
					fmt.Printf("%s\t%s\t-\tFIRST FAILED %v\n", a, b, err)

					continue
				}

				h2, _, err := bindPort(b, port)

				res := "OK"
				if err != nil {
					res = err.Error()
				} else {
					windows.Closesocket(h2)
				}

				windows.Closesocket(h1)
				fmt.Printf("%s\t%s\t%d\t%s\n", a, b, port, res)
			}
		}
	}

	// Does port 0 skip a port that another family holds explicitly? Bind
	// v6only explicitly to a port, then ask v4 for port 0 many times.
	fmt.Println()
}

// ---------------------------------------------------------------- collide

// collide allocates n port-0 sockets of a, then n of b, and counts ports
// that both sets hold at the same time. A shared allocation bitmap gives 0
// overlaps; independent pools give about n*n/rangeSize.
func collide(a, b spec, n int) {
	var ha, hb []windows.Handle

	pa, pb := map[int]bool{}, map[int]bool{}
	errA, errB := "", ""

	for range n {
		h, p, err := bindPort(a, 0)
		if err != nil {
			errA = err.Error()

			break
		}

		ha = append(ha, h)
		pa[p] = true
	}

	for range n {
		h, p, err := bindPort(b, 0)
		if err != nil {
			errB = err.Error()

			break
		}

		hb = append(hb, h)
		pb[p] = true
	}

	overlap := 0

	for p := range pb {
		if pa[p] {
			overlap++
		}
	}

	closeAll(ha)
	closeAll(hb)
	fmt.Printf("%s x %s\tn=%d\tgotA=%d\tgotB=%d\toverlap=%d\terrA=%q\terrB=%q\n", a, b, n, len(pa), len(pb), overlap, errA, errB)
}

func runCollide(n int) {
	fmt.Println("## collide: port 0 allocations held concurrently, overlap of port numbers")

	t4 := spec{fam: v4, proto: "tcp", addr: "wildcard"}
	t6 := spec{fam: v6only, proto: "tcp", addr: "wildcard"}
	td := spec{fam: dual, proto: "tcp", addr: "wildcard"}
	t4l := spec{fam: v4, proto: "tcp", addr: "loopback"}
	t4l2 := spec{fam: v4, proto: "tcp", addr: "loopback2"}
	t6l := spec{fam: v6only, proto: "tcp", addr: "loopback"}
	u4 := spec{fam: v4, proto: "udp", addr: "wildcard"}
	u6 := spec{fam: v6only, proto: "udp", addr: "wildcard"}

	collide(t4, t4, n)
	collide(t4, t6, n)
	collide(t6, t4, n)
	collide(t4, td, n)
	collide(t4l, t6l, n)
	collide(t4l, t4l2, n)
	collide(u4, u6, n)
	collide(t4, u4, n)
	collide(t6, u6, n)
}

// ---------------------------------------------------------------- exhaust

// fill allocates port-0 sockets of s until a bind fails or limit is reached.
func fill(s spec, limit int) ([]windows.Handle, []int, error) {
	var (
		hs    []windows.Handle
		ports []int
	)

	for range limit {
		h, p, err := bindPort(s, 0)
		if err != nil {
			return hs, ports, err
		}

		hs = append(hs, h)
		ports = append(ports, p)
	}

	return hs, ports, nil
}

func portSummary(ports []int) string {
	if len(ports) == 0 {
		return "none"
	}

	s := append([]int(nil), ports...)
	sort.Ints(s)

	return fmt.Sprintf("min=%d max=%d unique=%d", s[0], s[len(s)-1], uniq(s))
}

func uniq(s []int) int {
	m := map[int]bool{}
	for _, v := range s {
		m[v] = true
	}

	return len(m)
}

func errStr(err error) string {
	if err == nil {
		return "no failure within limit"
	}

	return err.Error()
}

// exhaustThenTry fills the pool with first, then counts how many port-0
// sockets of each probe still get a port.
func exhaustThenTry(name string, first spec, probes ...spec) {
	hs, ports, err := fill(first, maxSocketsPerSeries)
	fmt.Printf("### %s\nfill %s: %d sockets (%s), stopped by: %s\n", name, first, len(hs), portSummary(ports), errStr(err))
	tableCounts(first.proto)

	for _, p := range probes {
		ph, pports, perr := fill(p, maxSocketsPerSeries)
		fmt.Printf("  then %s: %d sockets (%s), stopped by: %s\n", p, len(ph), portSummary(pports), errStr(perr))
		closeAll(ph)
	}

	closeAll(hs)
}

func runExhaust() {
	t4 := spec{fam: v4, proto: "tcp", addr: "wildcard"}
	t6 := spec{fam: v6only, proto: "tcp", addr: "wildcard"}
	td := spec{fam: dual, proto: "tcp", addr: "wildcard"}
	t4l := spec{fam: v4, proto: "tcp", addr: "loopback"}
	t4l2 := spec{fam: v4, proto: "tcp", addr: "loopback2"}
	t6l := spec{fam: v6only, proto: "tcp", addr: "loopback"}
	u4 := spec{fam: v4, proto: "udp", addr: "wildcard"}
	u6 := spec{fam: v6only, proto: "udp", addr: "wildcard"}
	ud := spec{fam: dual, proto: "udp", addr: "wildcard"}

	fmt.Println("## exhaust")
	exhaustThenTry("E0 baseline capacity v4 tcp", t4)
	exhaustThenTry("E0 baseline capacity v6only tcp", t6)
	exhaustThenTry("E0 baseline capacity v4 udp", u4)
	exhaustThenTry("E1 v4 tcp exhausts, then other families/protocols", t4, t6, td, u4)
	exhaustThenTry("E2 v6only tcp exhausts", t6, t4, td)
	exhaustThenTry("E3 dual tcp exhausts", td, t4, t6)
	exhaustThenTry("E4 v4 tcp 127.0.0.1 exhausts", t4l, t4l2, t4, t6l)
	exhaustThenTry("E5 v4 udp exhausts", u4, u6, ud, t4)
	exhaustThenTry("E5b v6only udp exhausts", u6, u4)
	exhaustThenTry("E6 listening v4 tcp exhausts", spec{fam: v4, proto: "tcp", addr: "wildcard", listen: true}, t6, t4)
	exhaustThenTry("E6b SO_PORT_SCALABILITY v4 tcp 127.0.0.1", spec{fam: v4, proto: "tcp", addr: "loopback", scalable: true}, spec{fam: v4, proto: "tcp", addr: "loopback2", scalable: true}, t4)

	explicitBinds(t4, t4, "E7 explicit v4 binds inside range, then v4 port 0")
	explicitBinds(t6, t4, "E7b explicit v6only binds inside range, then v4 port 0")
	explicitBinds(u4, t4, "E7c explicit v4 udp binds inside range, then v4 tcp port 0")
	explicitBinds(t4l, t4, "E7d explicit v4 127.0.0.1 binds, then v4 wildcard port 0")
	explicitBinds(t4l, t4l2, "E7e explicit v4 127.0.0.1 binds, then v4 127.0.0.2 port 0")
}

//nolint:gochecknoglobals
var rangeStart, rangeCount int

func explicitBinds(s, probe spec, name string) {
	var hs []windows.Handle

	n := 100
	for p := rangeStart; p < rangeStart+n; p++ {
		h, _, err := bindPort(s, p)
		if err != nil {
			fmt.Printf("  explicit %s port %d: %v\n", s, p, err)

			continue
		}

		hs = append(hs, h)
	}

	ph, pports, perr := fill(probe, maxSocketsPerSeries)

	inExplicit := 0

	for _, p := range pports {
		if p >= rangeStart && p < rangeStart+n {
			inExplicit++
		}
	}

	fmt.Printf("### %s\nexplicit %s on %d..%d: %d held; then %s: %d sockets (%s), %d of them inside the explicit ports, stopped by: %s\n",
		name, s, rangeStart, rangeStart+n-1, len(hs), probe, len(ph), portSummary(pports), inExplicit, errStr(perr))
	closeAll(ph)
	closeAll(hs)
}

// ---------------------------------------------------------------- connect

func listener(addr string, closeOnAccept bool) (net.Listener, func()) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("listen", addr, err)
		os.Exit(1)
	}

	var (
		mu    sync.Mutex
		conns []net.Conn
	)

	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}

			if closeOnAccept {
				go func() {
					_, _ = io.Copy(io.Discard, c) // wait for the client FIN
					c.Close()
				}()

				continue
			}

			mu.Lock()
			conns = append(conns, c)
			mu.Unlock()
		}
	}()

	return l, func() {
		l.Close()
		mu.Lock()

		for _, c := range conns {
			c.Close()
		}

		mu.Unlock()
	}
}

func connectMany(s spec, remote windows.Sockaddr, limit int, bindFirst bool) ([]windows.Handle, []int, error) {
	var (
		hs    []windows.Handle
		ports []int
	)

	for range limit {
		h, err := s.open()
		if err != nil {
			return hs, ports, err
		}

		if bindFirst {
			if err := windows.Bind(h, s.sockaddr(0)); err != nil {
				windows.Closesocket(h)

				return hs, ports, fmt.Errorf("bind: %w", errName(err))
			}
		}

		if err := windows.Connect(h, remote); err != nil {
			windows.Closesocket(h)

			return hs, ports, fmt.Errorf("connect: %w", errName(err))
		}

		hs = append(hs, h)
		ports = append(ports, localPort(h))
	}

	return hs, ports, nil
}

func runConnect() {
	fmt.Println("## connect: implicit bind on connect()")

	_, stop1 := listener("127.0.0.1:40001", false)
	_, stop2 := listener("127.0.0.1:40002", false)
	_, stop6 := listener("[::1]:40003", false)

	r1 := &windows.SockaddrInet4{Port: 40001, Addr: [4]byte{127, 0, 0, 1}}
	r2 := &windows.SockaddrInet4{Port: 40002, Addr: [4]byte{127, 0, 0, 1}}
	r6 := &windows.SockaddrInet6{Port: 40003}
	r6.Addr[15] = 1

	c4 := spec{fam: v4, proto: "tcp", addr: "wildcard"}
	c6 := spec{fam: v6only, proto: "tcp", addr: "wildcard"}

	for _, unicast := range []bool{false, true} {
		a, b := c4, c6
		a.unicast, b.unicast = unicast, unicast

		h1, p1, e1 := connectMany(a, r1, maxSocketsPerSeries, false)
		fmt.Printf("### C1 unicast=%v connect v4 -> 127.0.0.1:40001: %d (%s), stopped by: %s\n", unicast, len(h1), portSummary(p1), errStr(e1))
		tableCounts("tcp")

		h2, p2, e2 := connectMany(a, r2, maxSocketsPerSeries, false)
		fmt.Printf("  then v4 -> 127.0.0.1:40002 (other remote port): %d (%s), reused ports of first set: %d, stopped by: %s\n", len(h2), portSummary(p2), overlap(p1, p2), errStr(e2))

		h3, p3, e3 := connectMany(b, r6, maxSocketsPerSeries, false)
		fmt.Printf("  then v6only -> [::1]:40003: %d (%s), reused ports of first set: %d, stopped by: %s\n", len(h3), portSummary(p3), overlap(p1, p3), errStr(e3))

		hb, pb, eb := fill(c4, maxSocketsPerSeries)
		fmt.Printf("  then v4 bind port 0: %d (%s), stopped by: %s\n", len(hb), portSummary(pb), errStr(eb))

		closeAll(hb)
		closeAll(h3)
		closeAll(h2)
		closeAll(h1)
		time.Sleep(time.Second)
	}

	stop1()
	stop2()
	stop6()

	// TIME_WAIT: the client closes first, the server closes after EOF.
	_, stopT := listener("127.0.0.1:40004", true)
	rt := &windows.SockaddrInet4{Port: 40004, Addr: [4]byte{127, 0, 0, 1}}

	hs, ps, err := connectMany(c4, rt, rangeCount/2, false)
	fmt.Printf("### C2 TIME_WAIT: %d connections (%s), err: %s\n", len(hs), portSummary(ps), errStr(err))
	closeAll(hs)
	time.Sleep(2 * time.Second)
	tableCounts("tcp")

	hb, pb, eb := fill(c4, maxSocketsPerSeries)
	fmt.Printf("  then v4 bind port 0 while %d sockets in TIME_WAIT: %d (%s), inside TIME_WAIT ports: %d, stopped by: %s\n", len(ps), len(hb), portSummary(pb), overlap(ps, pb), errStr(eb))
	closeAll(hb)

	hc, pc, ec := connectMany(c4, rt, maxSocketsPerSeries, false)
	fmt.Printf("  then v4 connect to the same remote: %d (%s), inside TIME_WAIT ports: %d, stopped by: %s\n", len(hc), portSummary(pc), overlap(ps, pc), errStr(ec))
	closeAll(hc)
	stopT()
}

func overlap(a, b []int) int {
	m := map[int]bool{}
	for _, p := range a {
		m[p] = true
	}

	n := 0

	for _, p := range b {
		if m[p] {
			n++
		}
	}

	return n
}

// ---------------------------------------------------------------- IP Helper

//nolint:gochecknoglobals
var (
	modiphlpapi             = windows.NewLazySystemDLL("iphlpapi.dll")
	procGetExtendedTcpTable = modiphlpapi.NewProc("GetExtendedTcpTable")
	procGetExtendedUdpTable = modiphlpapi.NewProc("GetExtendedUdpTable")
)

// tableCounts prints how many rows of the IP Helper tables have a local port
// inside the dynamic range, per state, so they can be compared with the
// number of sockets that actually hold a port.
func tableCounts(proto string) {
	for _, af := range []uint32{windows.AF_INET, windows.AF_INET6} {
		var rows map[string]int
		if proto == "tcp" {
			rows = tcpRows(af)
		} else {
			rows = udpRows(af)
		}

		fmt.Printf("    iphlpapi %s af=%d rows in range by state: %v\n", proto, af, rows)
	}
}

func tcpRows(af uint32) map[string]int {
	// TCP_TABLE_OWNER_PID_ALL = 5.
	buf := getTable(procGetExtendedTcpTable, af, 5)
	if buf == nil {
		return nil
	}

	n := binary.LittleEndian.Uint32(buf)
	out := map[string]int{}
	states := map[uint32]string{1: "CLOSED", 2: "LISTEN", 3: "SYN_SENT", 4: "SYN_RCVD", 5: "ESTAB", 6: "FIN_WAIT1", 7: "FIN_WAIT2", 8: "CLOSE_WAIT", 9: "CLOSING", 10: "LAST_ACK", 11: "TIME_WAIT", 12: "DELETE_TCB"}

	rowSize, portOff, stateOff := 24, 8, 0
	if af == windows.AF_INET6 {
		// MIB_TCP6ROW_OWNER_PID: local addr 16, scope 4, port 4, remote 16, scope 4, port 4, state 4, pid 4.
		rowSize, portOff, stateOff = 56, 20, 48
	}

	for i := range int(n) {
		r := buf[4+i*rowSize:]
		port := int(binary.BigEndian.Uint16(r[portOff:]))

		if port >= rangeStart && port < rangeStart+rangeCount {
			out[states[binary.LittleEndian.Uint32(r[stateOff:])]]++
		}
	}

	return out
}

func udpRows(af uint32) map[string]int {
	// UDP_TABLE_OWNER_PID = 1.
	buf := getTable(procGetExtendedUdpTable, af, 1)
	if buf == nil {
		return nil
	}

	n := binary.LittleEndian.Uint32(buf)
	rowSize, portOff := 12, 4

	if af == windows.AF_INET6 {
		rowSize, portOff = 28, 20
	}

	out := map[string]int{}

	for i := range int(n) {
		port := int(binary.BigEndian.Uint16(buf[4+i*rowSize+portOff:]))
		if port >= rangeStart && port < rangeStart+rangeCount {
			out["BOUND"]++
		}
	}

	return out
}

func getTable(p *windows.LazyProc, af, class uint32) []byte {
	size := uint32(0)
	_, _, _ = syscall.SyscallN(p.Addr(), 0, uintptr(unsafe.Pointer(&size)), 0, uintptr(af), uintptr(class), 0)

	for range 5 {
		size += 4096
		buf := make([]byte, size)

		r, _, _ := syscall.SyscallN(p.Addr(), uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&size)), 0, uintptr(af), uintptr(class), 0)
		if r == 0 {
			return buf
		}

		if r != uintptr(windows.ERROR_INSUFFICIENT_BUFFER) {
			fmt.Println("    table error", r)

			return nil
		}
	}

	return nil
}

// ---------------------------------------------------------------- NSI

//nolint:gochecknoglobals
var (
	modnsi                     = windows.NewLazySystemDLL("nsi.dll")
	procNsiGetAllParameters    = modnsi.NewProc("NsiGetAllParameters")
	procNsiGetParameter        = modnsi.NewProc("NsiGetParameter")
	procNsiAllocateAndGetTable = modnsi.NewProc("NsiAllocateAndGetTable")
	procNsiFreeTable           = modnsi.NewProc("NsiFreeTable")
)

type npiModuleID struct {
	length uint16
	_      uint16
	idType uint32
	guid   windows.GUID
}

func module(d1 uint32) *npiModuleID {
	return &npiModuleID{
		length: 24,
		idType: 1,
		guid:   windows.GUID{Data1: d1, Data2: 0x9b1a, Data3: 0x11d4, Data4: [8]byte{0x91, 0x23, 0x00, 0x50, 0x04, 0x77, 0x59, 0xbc}},
	}
}

//nolint:gochecknoglobals
var (
	tcpModule = module(0xeb004a03)
	udpModule = module(0xeb004a02)
)

func getAll(store uint32, m *npiModuleID, obj uint32, rwLen, dynLen, staticLen int) (rw, dyn, st []byte, err error) {
	rw, dyn, st = make([]byte, rwLen+1), make([]byte, dynLen+1), make([]byte, staticLen+1)

	ptr := func(b []byte, n int) (uintptr, uintptr) {
		if n == 0 {
			return 0, 0
		}

		return uintptr(unsafe.Pointer(&b[0])), uintptr(n)
	}

	rp, rl := ptr(rw, rwLen)
	dp, dl := ptr(dyn, dynLen)
	sp, sl := ptr(st, staticLen)

	r0, _, _ := syscall.SyscallN(procNsiGetAllParameters.Addr(), uintptr(store), uintptr(unsafe.Pointer(m)), uintptr(obj), 0, 0, rp, rl, dp, dl, sp, sl)
	if uint32(r0) != 0 {
		return nil, nil, nil, windows.Errno(uint32(r0))
	}

	return rw[:rwLen], dyn[:dynLen], st[:staticLen], nil
}

func getParam(store uint32, m *npiModuleID, obj, structType uint32, n, off int) ([]byte, error) {
	b := make([]byte, n)
	r0, _, _ := syscall.SyscallN(procNsiGetParameter.Addr(), uintptr(store), uintptr(unsafe.Pointer(m)), uintptr(obj), 0, 0, uintptr(structType), uintptr(unsafe.Pointer(&b[0])), uintptr(n), uintptr(off))

	if uint32(r0) != 0 {
		return nil, windows.Errno(uint32(r0))
	}

	return b, nil
}

// table enumerates a keyed object with NsiAllocateAndGetTable. Wine's
// prototype: (DWORD unk, const NPI_MODULEID*, DWORD table, void **key,
// DWORD key_size, void **rw, DWORD rw_size, void **dyn, DWORD dyn_size,
// void **static, DWORD static_size, DWORD *count, DWORD unk2).
func table(store uint32, m *npiModuleID, obj uint32, keySize, rwSize int) ([][]byte, [][]byte, error) {
	var keys, rws, dyn, st uintptr

	count := uint32(0)

	rwPtr, rwLen := uintptr(unsafe.Pointer(&rws)), uintptr(rwSize)
	if rwSize == 0 {
		rwPtr, rwLen = 0, 0
	}

	r0, _, _ := syscall.SyscallN(procNsiAllocateAndGetTable.Addr(),
		uintptr(store), uintptr(unsafe.Pointer(m)), uintptr(obj),
		uintptr(unsafe.Pointer(&keys)), uintptr(keySize),
		rwPtr, rwLen,
		0, 0, // no dynamic part
		0, 0, // no static part
		uintptr(unsafe.Pointer(&count)), 0)
	if uint32(r0) != 0 {
		return nil, nil, windows.Errno(uint32(r0))
	}

	var ks, rs [][]byte

	for i := range int(count) {
		ks = append(ks, append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(keys+uintptr(i*keySize))), keySize)...))
		if rws != 0 {
			rs = append(rs, append([]byte(nil), unsafe.Slice((*byte)(unsafe.Pointer(rws+uintptr(i*rwSize))), rwSize)...))
		}
	}

	_, _, _ = syscall.SyscallN(procNsiFreeTable.Addr(), keys, rws, dyn, st)

	return ks, rs, nil
}

func runNSI() {
	fmt.Println("## nsi")

	for _, o := range []struct {
		name string
		m    *npiModuleID
		obj  uint32
		rw   int
	}{
		{"tcp range (20)", tcpModule, 20, 4},
		{"udp range (4)", udpModule, 4, 4},
		{"tcp portpool ex (34)", tcpModule, 34, 44},
	} {
		for _, store := range []uint32{0, 1, 3} {
			rw, _, _, err := getAll(store, o.m, o.obj, o.rw, 0, 0)
			fmt.Printf("getall store=%d %s: %s err=%v\n", store, o.name, hex.EncodeToString(rw), err)
		}

		b, err := getParam(1, o.m, o.obj, 0, o.rw, 0)
		fmt.Printf("getparam store=1 %s: %s err=%v\n", o.name, hex.EncodeToString(b), err)
	}

	// Wrong sizes: what does the kernel return?
	_, _, _, err := getAll(1, tcpModule, 34, 40, 0, 0)
	fmt.Printf("getall tcp 34 with 40 bytes: err=%v\n", err)
	_, _, _, err = getAll(1, tcpModule, 20, 8, 0, 0)
	fmt.Printf("getall tcp 20 with 8 bytes: err=%v\n", err)
	_, _, _, err = getAll(1, tcpModule, 99, 4, 0, 0)
	fmt.Printf("getall tcp 99: err=%v\n", err)

	for _, o := range []struct {
		name string
		m    *npiModuleID
		obj  uint32
	}{
		{"tcp exclusions (29)", tcpModule, 29},
		{"udp exclusions (5)", udpModule, 5},
	} {
		for _, store := range []uint32{0, 1} {
			for _, rw := range []int{2, 0} {
				ks, rs, err := table(store, o.m, o.obj, 4, rw)
				fmt.Printf("table store=%d rw=%d %s: n=%d err=%v\n", store, rw, o.name, len(ks), err)

				for i, k := range ks {
					r := ""
					if i < len(rs) {
						r = hex.EncodeToString(rs[i])
					}

					fmt.Printf("  key=%s start=%d count=%d rw=%s\n", hex.EncodeToString(k), binary.LittleEndian.Uint16(k), binary.LittleEndian.Uint16(k[2:]), r)
				}
			}
		}
	}

	// Tables near the known ones, to look for reservation objects or counters.
	for _, m := range []struct {
		name string
		m    *npiModuleID
	}{{"tcp", tcpModule}, {"udp", udpModule}} {
		for obj := uint32(0); obj < 48; obj++ {
			_, _, _, err := getAll(1, m.m, obj, 0, 0, 0)
			fmt.Printf("probe getall %s obj=%d no buffers: err=%v\n", m.name, obj, err)
		}
	}
}

// runObj34 dumps TCP object 34 (port pool ex) and the range object 20 from
// stores 1 and 3, with the first 4 bytes decoded like object 20 and the rest
// as little-endian uint16 and uint32 words.
func runObj34() {
	for _, store := range []uint32{1, 3} {
		r20, _, _, err := getAll(store, tcpModule, 20, 4, 0, 0)
		fmt.Printf("obj34: store=%d tcp 20: %s err=%v\n", store, hex.EncodeToString(r20), err)

		b, _, _, err := getAll(store, tcpModule, 34, 44, 0, 0)
		fmt.Printf("obj34: store=%d tcp 34: %s err=%v\n", store, hex.EncodeToString(b), err)

		if len(b) < 44 {
			continue
		}

		fmt.Printf("obj34:   smart start=%d count=%d\n", binary.BigEndian.Uint16(b), binary.LittleEndian.Uint16(b[2:]))

		var u16, u32 []string
		for i := 4; i+2 <= len(b); i += 2 {
			u16 = append(u16, fmt.Sprintf("+%d=%d", i, binary.LittleEndian.Uint16(b[i:])))
		}

		for i := 4; i+4 <= len(b); i += 4 {
			u32 = append(u32, fmt.Sprintf("+%d=%d", i, binary.LittleEndian.Uint32(b[i:])))
		}

		fmt.Printf("obj34:   u16 %s\n", strings.Join(u16, " "))
		fmt.Printf("obj34:   u32 %s\n", strings.Join(u32, " "))
	}

	for _, o := range []struct {
		name string
		m    *npiModuleID
		obj  uint32
	}{{"tcp exclusions (29)", tcpModule, 29}, {"udp exclusions (5)", udpModule, 5}} {
		ks, _, err := table(1, o.m, o.obj, 4, 2)
		fmt.Printf("obj34: %s n=%d err=%v\n", o.name, len(ks), err)

		for _, k := range ks {
			fmt.Printf("obj34:   start=%d count=%d\n", binary.LittleEndian.Uint16(k), binary.LittleEndian.Uint16(k[2:]))
		}
	}
}

// ---------------------------------------------------------------- hold

//nolint:gochecknoglobals
var (
	procInternalGetBoundTcpEndpointTable  = modiphlpapi.NewProc("InternalGetBoundTcpEndpointTable")
	procInternalGetBoundTcp6EndpointTable = modiphlpapi.NewProc("InternalGetBoundTcp6EndpointTable")
	procInternalGetTcpDynamicPortRange    = modiphlpapi.NewProc("InternalGetTcpDynamicPortRange")
	procInternalGetUdpDynamicPortRange    = modiphlpapi.NewProc("InternalGetUdpDynamicPortRange")
	modkernel32                           = windows.NewLazySystemDLL("kernel32.dll")
	procGetProcessHeap                    = modkernel32.NewProc("GetProcessHeap")
	procHeapSize                          = modkernel32.NewProc("HeapSize")
	procHeapFree                          = modkernel32.NewProc("HeapFree")
)

// boundTable calls the undocumented iphlpapi export and dumps what it returns.
func boundTable(p *windows.LazyProc, name string) {
	if err := p.Find(); err != nil {
		fmt.Printf("%s: %v\n", name, err)

		return
	}

	heap, _, _ := procGetProcessHeap.Call()

	var tbl uintptr

	r0, _, _ := syscall.SyscallN(p.Addr(), uintptr(unsafe.Pointer(&tbl)), heap, 0)
	if r0 != 0 || tbl == 0 {
		fmt.Printf("%s: error %d\n", name, r0)

		return
	}

	size, _, _ := syscall.SyscallN(procHeapSize.Addr(), heap, 0, tbl)
	b := unsafe.Slice((*byte)(unsafe.Pointer(tbl)), int(size))
	n := binary.LittleEndian.Uint32(b)

	inRange := 0

	if n > 0 {
		rowSize := (int(size) - 4) / int(n)
		fmt.Printf("%s: n=%d heap=%d bytes (approx row %d)\n  first rows: %s\n", name, n, size, rowSize, hex.EncodeToString(b[:min(len(b), 4+3*rowSize+8)]))
	} else {
		fmt.Printf("%s: n=0 heap=%d\n", name, size)
	}

	_ = inRange
	_, _, _ = syscall.SyscallN(procHeapFree.Addr(), heap, 0, tbl)
}

func iphlpRange() {
	for _, p := range []*windows.LazyProc{procInternalGetTcpDynamicPortRange, procInternalGetUdpDynamicPortRange} {
		if err := p.Find(); err != nil {
			fmt.Println(p.Name, err)

			continue
		}

		var start, count uint16
		r0, _, _ := syscall.SyscallN(p.Addr(), uintptr(unsafe.Pointer(&start)), uintptr(unsafe.Pointer(&count)))
		fmt.Printf("%s: r=%d start=%d (0x%04x) count=%d (0x%04x)\n", p.Name, r0, start, start, count, count)
	}
}

// runHold holds 37 bound-only, 23 listening and 11 connected TCP sockets
// and shows which tables list them.
func runHold() {
	_, stop := listener("127.0.0.1:40011", false)
	defer stop()

	var hs []windows.Handle

	b, bp, _ := fill(spec{fam: v4, proto: "tcp", addr: "wildcard"}, 37)
	l, lp, _ := fill(spec{fam: v4, proto: "tcp", addr: "wildcard", listen: true}, 23)
	c, cp, _ := connectMany(spec{fam: v4, proto: "tcp", addr: "wildcard"}, &windows.SockaddrInet4{Port: 40011, Addr: [4]byte{127, 0, 0, 1}}, 11, false)
	b6, _, _ := fill(spec{fam: v6only, proto: "tcp", addr: "wildcard"}, 5)
	hs = append(append(append(append(hs, b...), l...), c...), b6...)

	fmt.Printf("## hold: bound %v\n  listen %v\n  connected %v\n", bp, lp, cp)
	tableCounts("tcp")
	boundTable(procInternalGetBoundTcpEndpointTable, "InternalGetBoundTcpEndpointTable")
	boundTable(procInternalGetBoundTcp6EndpointTable, "InternalGetBoundTcp6EndpointTable")
	iphlpRange()

	out, _ := exec.Command("netstat", "-anq", "-p", "TCP").CombinedOutput()
	bound := 0

	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "BOUND") {
			bound++
		}
	}

	fmt.Printf("netstat -anq BOUND rows (tcp v4): %d\n", bound)

	closeAll(hs)
}

// runUnicast checks SO_REUSE_UNICASTPORT with fresh listeners: can one local
// port serve connections to different remote endpoints?
func runUnicast() {
	for i, mode := range []struct {
		name      string
		unicast   bool
		bindFirst bool
	}{{"unicast connect", true, false}, {"unicast bind(0)+connect", true, true}, {"plain connect", false, false}} {
		// Fresh listeners per mode, so FIN_WAIT_2 leftovers of an earlier
		// mode cannot collide with the 4-tuples of this one.
		p1, p2 := 40100+i*2, 40101+i*2
		_, s1 := listener(fmt.Sprintf("127.0.0.1:%d", p1), false)
		_, s2 := listener(fmt.Sprintf("127.0.0.1:%d", p2), false)
		r1 := &windows.SockaddrInet4{Port: p1, Addr: [4]byte{127, 0, 0, 1}}
		r2 := &windows.SockaddrInet4{Port: p2, Addr: [4]byte{127, 0, 0, 1}}

		s := spec{fam: v4, proto: "tcp", addr: "wildcard", unicast: mode.unicast}
		h1, ports1, e1 := connectMany(s, r1, maxSocketsPerSeries, mode.bindFirst)
		h2, ports2, e2 := connectMany(s, r2, maxSocketsPerSeries, mode.bindFirst)
		fmt.Printf("### %s: to :%d %d (%s) stop=%s; to :%d %d (%s) shared ports=%d stop=%s\n",
			mode.name, p1, len(h1), portSummary(ports1), errStr(e1), p2, len(h2), portSummary(ports2), overlap(ports1, ports2), errStr(e2))
		closeAll(h1)
		closeAll(h2)
		s1()
		s2()
		time.Sleep(3 * time.Second)
	}
}

// runBench reads the range and the exclusion tables concurrently and in a
// loop, and checks that NsiFreeTable releases what NsiAllocateAndGetTable
// allocated.
func runBench() {
	const iterations = 200000

	// PROCESS_MEMORY_COUNTERS: cb, PageFaultCount, then 8 SIZE_T fields;
	// PagefileUsage is the 7th of them.
	var pmc struct {
		cb, faults uint32
		sizes      [8]uintptr
	}

	proc := modkernel32.NewProc("K32GetProcessMemoryInfo")
	mem := func() uintptr {
		pmc.cb = uint32(unsafe.Sizeof(pmc))
		_, _, _ = syscall.SyscallN(proc.Addr(), uintptr(windows.CurrentProcess()), uintptr(unsafe.Pointer(&pmc)), uintptr(pmc.cb))

		return pmc.sizes[6]
	}

	for _, kind := range []string{"range", "table", "table", "table"} {
		before := mem()
		start := time.Now()

		var (
			wg   sync.WaitGroup
			mu   sync.Mutex
			errs = map[string]int{}
		)

		for range 8 {
			wg.Add(1)

			go func() {
				defer wg.Done()

				for range iterations / 8 {
					var err error
					if kind == "range" {
						_, _, _, err = getAll(1, tcpModule, 20, 4, 0, 0)
					} else {
						_, _, err = table(1, tcpModule, 29, 4, 2)
					}

					if err != nil {
						mu.Lock()
						errs[err.Error()]++
						mu.Unlock()
					}
				}
			}()
		}

		wg.Wait()
		el := time.Since(start)
		runtime.GC()
		debug.FreeOSMemory()
		fmt.Printf("bench %s: %d calls on 8 goroutines in %v (%.1f us/call), errors=%v, pagefile usage %d -> %d KiB\n",
			kind, iterations, el, float64(el.Microseconds())/iterations, errs, before/1024, mem()/1024)
	}
}

// ---------------------------------------------------------------- validate

// boundPorts returns the local ports of TCP endpoints that are bound but not
// listening or connected (state 100 in MIB_TCPROW2/MIB_TCP6ROW2), read with
// the undocumented iphlpapi exports netstat -q uses.
func boundPorts() []int {
	var out []int

	heap, _, _ := procGetProcessHeap.Call()

	for _, t := range []struct {
		p                *windows.LazyProc
		rowSize, portOff int
	}{{procInternalGetBoundTcpEndpointTable, 28, 8}, {procInternalGetBoundTcp6EndpointTable, 60, 20}} {
		var tbl uintptr

		r0, _, _ := syscall.SyscallN(t.p.Addr(), uintptr(unsafe.Pointer(&tbl)), heap, 0)
		if r0 != 0 || tbl == 0 {
			fmt.Println("bound table error", r0)

			continue
		}

		n := *(*uint32)(unsafe.Pointer(tbl))
		b := unsafe.Slice((*byte)(unsafe.Pointer(tbl)), 4+int(n)*t.rowSize)

		for i := range int(n) {
			out = append(out, int(binary.BigEndian.Uint16(b[4+i*t.rowSize+t.portOff:])))
		}

		_, _, _ = syscall.SyscallN(procHeapFree.Addr(), heap, 0, tbl)
	}

	return out
}

// tablePorts returns local ports of the IP Helper TCP (all states) or UDP
// tables for IPv4 and IPv6, with TIME_WAIT and FIN_WAIT2 rows separately.
// Both hold only a weak port reference after closesocket, so they do not
// block bind(0). A half-closed FIN_WAIT2 socket that is still open does hold
// its port; the table cannot tell the two apart.
func tablePorts(proto string) (ports, timeWait []int) {
	for _, af := range []uint32{windows.AF_INET, windows.AF_INET6} {
		if proto == "udp" {
			buf := getTable(procGetExtendedUdpTable, af, 1)
			rowSize, portOff := 12, 4

			if af == windows.AF_INET6 {
				rowSize, portOff = 28, 20
			}

			for i := range int(binary.LittleEndian.Uint32(buf)) {
				ports = append(ports, int(binary.BigEndian.Uint16(buf[4+i*rowSize+portOff:])))
			}

			continue
		}

		buf := getTable(procGetExtendedTcpTable, af, 5)
		rowSize, portOff, stateOff := 24, 8, 0

		if af == windows.AF_INET6 {
			rowSize, portOff, stateOff = 56, 20, 48
		}

		for i := range int(binary.LittleEndian.Uint32(buf)) {
			r := buf[4+i*rowSize:]
			p := int(binary.BigEndian.Uint16(r[portOff:]))

			if st := binary.LittleEndian.Uint32(r[stateOff:]); st == 11 || st == 7 {
				timeWait = append(timeWait, p)
			} else {
				ports = append(ports, p)
			}
		}
	}

	return ports, timeWait
}

// estimate predicts how many port-0 binds still succeed: range size minus the
// union of excluded ports and local ports of endpoints inside the range.
func estimate(proto string) (free, excluded, used, timeWaitOnly int) {
	inRange := func(p int) bool { return p >= rangeStart && p < rangeStart+rangeCount }
	set := map[int]bool{}

	m, obj := tcpModule, uint32(29)
	if proto == "udp" {
		m, obj = udpModule, 5
	}

	keys, _, err := table(1, m, obj, 4, 0)
	if err != nil {
		fmt.Println("exclusions:", err)
	}

	for _, k := range keys {
		start, count := int(binary.LittleEndian.Uint16(k)), int(binary.LittleEndian.Uint16(k[2:]))
		for p := start; p < start+count; p++ {
			if inRange(p) && !set[p] {
				set[p] = true
				excluded++
			}
		}
	}

	ports, tw := tablePorts(proto)
	if proto == "tcp" {
		ports = append(ports, boundPorts()...)
	}

	for _, p := range ports {
		if inRange(p) && !set[p] {
			set[p] = true
			used++
		}
	}

	for _, p := range tw {
		if inRange(p) && !set[p] {
			timeWaitOnly++
		}
	}

	return rangeCount - len(set), excluded, used, timeWaitOnly
}

func check(name, proto string) {
	free, excl, used, tw := estimate(proto)
	tableCounts(proto)
	sp := spec{fam: v4, proto: proto, addr: "wildcard"}
	hs, _, err := fill(sp, maxSocketsPerSeries)
	closeAll(hs)
	fmt.Printf("### %s %s: predicted free=%d (excluded=%d, endpoint ports=%d, TIME_WAIT-only ports=%d), actual port-0 binds=%d (%s), diff=%d\n",
		name, proto, free, excl, used, tw, len(hs), errStr(err), len(hs)-free)
}

// runValidate compares the estimate from public and semi-public tables with
// the measured capacity under different socket mixes.
func runValidate() {
	check("V0 idle", "tcp")
	check("V0 idle", "udp")

	_, stop := listener("127.0.0.1:40031", false)
	defer stop()

	var hs []windows.Handle

	add := func(h []windows.Handle, _ []int, _ error) { hs = append(hs, h...) }
	add(fill(spec{fam: v4, proto: "tcp", addr: "wildcard"}, 30))
	add(fill(spec{fam: v6only, proto: "tcp", addr: "wildcard", listen: true}, 20))
	add(fill(spec{fam: dual, proto: "tcp", addr: "wildcard"}, 10))
	add(fill(spec{fam: v4, proto: "tcp", addr: "loopback2"}, 10))
	add(connectMany(spec{fam: v4, proto: "tcp", addr: "wildcard"}, &windows.SockaddrInet4{Port: 40031, Addr: [4]byte{127, 0, 0, 1}}, 20, false))

	for p := rangeStart + rangeCount - 10; p < rangeStart+rangeCount; p++ {
		if h, _, err := bindPort(spec{fam: v6only, proto: "tcp", addr: "loopback"}, p); err == nil {
			hs = append(hs, h)
		}
	}

	add(fill(spec{fam: v4, proto: "udp", addr: "wildcard"}, 25))
	add(fill(spec{fam: v6only, proto: "udp", addr: "wildcard"}, 15))
	check("V1 mixed sockets held", "tcp")
	check("V1 mixed sockets held", "udp")
	closeAll(hs)
	time.Sleep(time.Second)

	_, stopT := listener("127.0.0.1:40032", true)
	tw, _, _ := connectMany(spec{fam: v4, proto: "tcp", addr: "wildcard"}, &windows.SockaddrInet4{Port: 40032, Addr: [4]byte{127, 0, 0, 1}}, 60, false)
	closeAll(tw)
	time.Sleep(2 * time.Second)
	check("V2 60 TIME_WAIT", "tcp")
	stopT()
}

// ---------------------------------------------------------------- restricted

// runRestricted drops to a token without the Administrators group and
// without privileges before the first NSI call of this process.
func runRestricted() {
	runtime.LockOSThread()

	var tok windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_DUPLICATE|windows.TOKEN_QUERY|windows.TOKEN_IMPERSONATE|windows.TOKEN_ASSIGN_PRIMARY, &tok); err != nil {
		fmt.Println("OpenProcessToken:", err)

		return
	}

	admins, _ := windows.CreateWellKnownSid(windows.WinBuiltinAdministratorsSid)
	deny := []windows.SIDAndAttributes{{Sid: admins}}

	var restricted windows.Token

	// DISABLE_MAX_PRIVILEGE = 1.
	r, _, e := modadvapi32.NewProc("CreateRestrictedToken").Call(uintptr(tok), 1, 1, uintptr(unsafe.Pointer(&deny[0])), 0, 0, 0, 0, uintptr(unsafe.Pointer(&restricted)))
	if r == 0 {
		fmt.Println("CreateRestrictedToken:", e)

		return
	}

	var imp windows.Token
	if err := windows.DuplicateTokenEx(restricted, windows.TOKEN_ALL_ACCESS, nil, windows.SecurityImpersonation, windows.TokenImpersonation, &imp); err != nil {
		fmt.Println("DuplicateTokenEx:", err)

		return
	}

	if err := windows.SetThreadToken(nil, imp); err != nil {
		fmt.Println("SetThreadToken:", err)

		return
	}

	var th windows.Token
	_ = windows.OpenThreadToken(windows.CurrentThread(), windows.TOKEN_QUERY, true, &th)
	member, _ := th.IsMember(admins)
	fmt.Printf("## restricted token: member of Administrators=%v elevated=%v\n", member, th.IsElevated())

	runNSIOnThread()
	boundTable(procInternalGetBoundTcpEndpointTable, "InternalGetBoundTcpEndpointTable")
	iphlpRange()
}

//nolint:gochecknoglobals
var modadvapi32 = windows.NewLazySystemDLL("advapi32.dll")

// runNSIOnThread is the read-only part of runNSI without the object scan.
func runNSIOnThread() {
	rw, _, _, err := getAll(1, tcpModule, 20, 4, 0, 0)
	fmt.Printf("getall tcp 20: %s err=%v\n", hex.EncodeToString(rw), err)
	rw, _, _, err = getAll(1, udpModule, 4, 4, 0, 0)
	fmt.Printf("getall udp 4: %s err=%v\n", hex.EncodeToString(rw), err)
	rw, _, _, err = getAll(1, tcpModule, 34, 44, 0, 0)
	fmt.Printf("getall tcp 34: %s err=%v\n", hex.EncodeToString(rw), err)

	for _, o := range []struct {
		m   *npiModuleID
		obj uint32
	}{{tcpModule, 29}, {udpModule, 5}} {
		ks, _, err := table(1, o.m, o.obj, 4, 2)
		fmt.Printf("table %d: n=%d err=%v\n", o.obj, len(ks), err)
	}
}

// ---------------------------------------------------------------- main

func netsh(args ...string) {
	out, err := exec.Command("netsh", args...).CombinedOutput()
	fmt.Printf("> netsh %s (err=%v)\n%s\n", strings.Join(args, " "), err, strings.TrimSpace(string(out)))
}

func main() {
	n := flag.Int("n", 1500, "sockets per side for collide")
	flag.IntVar(&rangeStart, "start", 49152, "dynamic range start (for statistics)")
	flag.IntVar(&rangeCount, "count", 16384, "dynamic range size (for statistics)")
	flag.Parse()

	switch flag.Arg(0) {
	case "nsi":
		runNSI()
	case "pairs":
		runPairs()
	case "collide":
		runCollide(*n)
	case "exhaust":
		runExhaust()
	case "connect":
		runConnect()
	case "hold":
		runHold()
	case "unicast":
		runUnicast()
	case "bench":
		runBench()
	case "restricted":
		runRestricted()
	case "validate":
		runValidate()
	case "capacity":
		for _, sp := range []spec{{fam: v4, proto: "tcp", addr: "wildcard"}, {fam: v6only, proto: "tcp", addr: "wildcard"}, {fam: v4, proto: "udp", addr: "wildcard"}, {fam: v6only, proto: "udp", addr: "wildcard"}} {
			hs, ps, err := fill(sp, maxSocketsPerSeries)
			fmt.Printf("capacity %s: %d (%s), stopped by: %s\n", sp, len(hs), portSummary(ps), errStr(err))
			closeAll(hs)
		}
	case "obj34":
		runObj34()
	case "netsh":
		netsh(flag.Args()[1:]...)
	default:
		fmt.Println("usage: portprobe nsi|pairs|collide|exhaust|connect")
		os.Exit(2)
	}
}
