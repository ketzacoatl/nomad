package nomad

import (
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/hashicorp/raft"
	"github.com/hashicorp/serf/serf"
)

const (
	DefaultRegion   = "region1"
	DefaultDC       = "dc1"
	DefaultSerfPort = 4647
)

// These are the protocol versions that Nomad can understand
const (
	ProtocolVersionMin uint8 = 1
	ProtocolVersionMax       = 1
)

// ProtocolVersionMap is the mapping of Nomad protocol versions
// to Serf protocol versions. We mask the Serf protocols using
// our own protocol version.
var protocolVersionMap map[uint8]uint8

func init() {
	protocolVersionMap = map[uint8]uint8{
		1: 5,
	}
}

var (
	DefaultRPCAddr = &net.TCPAddr{IP: net.ParseIP("0.0.0.0"), Port: 4646}
)

// Config is used to parameterize the server
type Config struct {
	// Bootstrap mode is used to bring up the first Consul server.
	// It is required so that it can elect a leader without any
	// other nodes being present
	Bootstrap bool

	// BootstrapExpect mode is used to automatically bring up a collection of
	// Consul servers. This can be used to automatically bring up a collection
	// of nodes.
	BootstrapExpect int

	// DataDir is the directory to store our state in
	DataDir string

	// DevMode is used for development purposes only and limits the
	// use of persistence or state.
	DevMode bool

	// DevDisableBootstrap is used to disable bootstrap mode while
	// in DevMode. This is largely used for testing.
	DevDisableBootstrap bool

	// LogOutput is the location to write logs to. If this is not set,
	// logs will go to stderr.
	LogOutput io.Writer

	// ProtocolVersion is the protocol version to speak. This must be between
	// ProtocolVersionMin and ProtocolVersionMax.
	ProtocolVersion uint8

	// RPCAddr is the RPC address used by Nomad. This should be reachable
	// by the other servers and clients
	RPCAddr *net.TCPAddr

	// RPCAdvertise is the address that is advertised to other nodes for
	// the RPC endpoint. This can differ from the RPC address, if for example
	// the RPCAddr is unspecified "0.0.0.0:4646", but this address must be
	// reachable
	RPCAdvertise *net.TCPAddr

	// RaftConfig is the configuration used for Raft in the local DC
	RaftConfig *raft.Config

	// RequireTLS ensures that all RPC traffic is protected with TLS
	RequireTLS bool

	// SerfConfig is the configuration for the serf cluster
	SerfConfig *serf.Config

	// Node name is the name we use to advertise. Defaults to hostname.
	NodeName string

	// Region is the region this Nomad server belongs to.
	Region string

	// Datacenter is the datacenter this Nomad server belongs to.
	Datacenter string

	// Build is a string that is gossiped around, and can be used to help
	// operators track which versions are actively deployed
	Build string

	// ReconcileInterval controls how often we reconcile the strongly
	// consistent store with the Serf info. This is used to handle nodes
	// that are force removed, as well as intermittent unavailability during
	// leader election.
	ReconcileInterval time.Duration
}

// CheckVersion is used to check if the ProtocolVersion is valid
func (c *Config) CheckVersion() error {
	if c.ProtocolVersion < ProtocolVersionMin {
		return fmt.Errorf("Protocol version '%d' too low. Must be in range: [%d, %d]",
			c.ProtocolVersion, ProtocolVersionMin, ProtocolVersionMax)
	} else if c.ProtocolVersion > ProtocolVersionMax {
		return fmt.Errorf("Protocol version '%d' too high. Must be in range: [%d, %d]",
			c.ProtocolVersion, ProtocolVersionMin, ProtocolVersionMax)
	}
	return nil
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	hostname, err := os.Hostname()
	if err != nil {
		panic(err)
	}

	c := &Config{
		Region:            DefaultRegion,
		Datacenter:        DefaultDC,
		NodeName:          hostname,
		ProtocolVersion:   ProtocolVersionMax,
		RaftConfig:        raft.DefaultConfig(),
		RPCAddr:           DefaultRPCAddr,
		SerfConfig:        serf.DefaultConfig(),
		ReconcileInterval: 60 * time.Second,
	}

	// Serf should use the WAN timing, since we are using it
	// to communicate between DC's
	c.SerfConfig.MemberlistConfig = memberlist.DefaultWANConfig()
	c.SerfConfig.MemberlistConfig.BindPort = DefaultSerfPort

	// Disable shutdown on removal
	c.RaftConfig.ShutdownOnRemove = false
	return c
}