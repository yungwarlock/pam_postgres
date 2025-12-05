package requestaccess

import (
	"fmt"
	"net/netip"
	"os"
	"pam_postgres/pkg/etchosts"
	dbnet "pam_postgres/services/db_net"
)

var (
	Debug = os.Getenv("DEBUG") == "1"
)

func approveAccessRequest() string {
	name, port, fullName := dbnet.GenerateSubdomainAndPort()
	if Debug {
		fmt.Printf("Debug: Generated subdomain %s and port %s\n", fullName, port)
		etchosts.Add("/etc/hosts", []etchosts.Record{{
			Hosts: fullName,
			IP:    netip.MustParseAddr("127.0.0.1"),
		}})
	}

	go dbnet.CreateConnection(name, port)
	fmt.Printf("Access approved. Connect to %s on port %s\n", fullName, port)
	return fullName
}
