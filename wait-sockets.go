package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strings"
	"time"
)

const (
	overallTimeout  = 30 * time.Second
	perTryTimeout   = 2 * time.Second
	numConnAttempts = int(overallTimeout / perTryTimeout)
)

// docker-compose passes the locations of sockets that the current
// container cares about in environmental variables.  Here we wait
// until we can establish TCP sockets to the addresses specified, a
// proxy for whether our dependencies are ready for business.
func WaitSockets() error {
	env := os.Environ()
	sort.Strings(env)

	addrs := make(map[string]struct{})

	for _, pair := range env {
		pieces := strings.SplitN(pair, "=", 2)
		if len(pieces) != 2 {
			log.Printf("bad environmental var: %s", pair)
			continue
		}
		k, v := pieces[0], pieces[1]
		if !strings.HasSuffix(k, "_PORT") || strings.Contains(k, "TCP") || strings.Contains(k, "UDP") {
			continue
		}
		if !strings.HasPrefix(v, "tcp://") {
			continue
		}
		addr := v[len("tcp://"):]

		// docker-compose can provide the same address in
		// multiple vars. take the easy way out and just
		// deduplicate here.
		addrs[addr] = struct{}{}
	}

	for addr, _ := range addrs {
		var err error
		var i int
		for i = 0; i < numConnAttempts; i++ {
			var conn net.Conn
			start := time.Now()
			conn, err = net.DialTimeout("tcp", addr, 2*time.Second)
			if err == nil {
				conn.Close()
				break
			}
			// if we get a ConnRefused error, the
			// DialTimeout might return immediately.  If
			// we don't sleep the rest of our timeout, we
			// will blow through all 15 connection
			// attempts in a matter of milliseconds.
			time.Sleep(start.Add(perTryTimeout).Sub(time.Now()))
		}
		if err != nil {
			return fmt.Errorf("DialTimeout(%s) after %d attempts: %s", addr, i, err)
		}
	}

	return nil
}
