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
		for i := 0; i < numConnAttempts; i++ {
			var conn net.Conn
			conn, err = net.DialTimeout("tcp", addr, 2*time.Second)
			if err == nil {
				conn.Close()
				break
			}
		}
		if err != nil {
			return fmt.Errorf("DialTimeout(%s)", addr, err)
		}
	}

	return nil
}
