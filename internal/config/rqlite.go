package config

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// All variables will be set, unless explicit noted.
// TODO: set sensible defaults
type RqLite struct {
	// DataPath is path to node data. Always set.
	DataPath string `yaml:"dataPath" json:"dataPath" env:"RQLITE_DATA_PATH"`

	// HTTPAddr is the bind network address for the HTTP Server.
	// It never includes a trailing HTTP or HTTPS.
	HTTPAddr string `yaml:"httpAddr" json:"httpAddr" env:"RQLITE_HTTP_ADDR"`

	// HTTPAdv is the advertised HTTP server network.
	HTTPAdv string `yaml:"httpAdv" json:"httpAdv" env:"RQLITE_HTTP_ADV"`

	// HTTPAllowOrigin is the value to set for Access-Control-Allow-Origin HTTP header.
	HTTPAllowOrigin string `yaml:"httpAllowOrigin" json:"httpAllowOrigin" env:"RQLITE_HTTP_ALLOW_ORIGIN"`

	// AuthFile is the path to the authentication file. May not be set.
	AuthFile string `yaml:"authFile" json:"authFile" env:"RQLITE_AUTH_FILE"`

	// AutoBackupFile is the path to the auto-backup file. May not be set.
	AutoBackupFile string `filepath:"true" yaml:"autoBackupFile" json:"autoBackupFile" env:"RQLITE_AUTO_BACKUP_FILE"`

	// AutoRestoreFile is the path to the auto-restore file. May not be set.
	AutoRestoreFile string `filepath:"true" yaml:"autoRestoreFile" json:"autoRestoreFile" env:"RQLITE_AUTO_RESTORE_FILE"`

	// HTTPx509CACert is the path to the CA certificate file for when this node verifies
	// other certificates for any HTTP communications. May not be set.
	HTTPx509CACert string `filepath:"true" yaml:"httpx509CACert" json:"httpx509CACert" env:"RQLITE_HTTP_X509_CA_CERT"`

	// HTTPx509Cert is the path to the X509 cert for the HTTP server. May not be set.
	HTTPx509Cert string `filepath:"true" yaml:"httpx509Cert" json:"httpx509Cert" env:"RQLITE_HTTP_X509_CERT"`

	// HTTPx509Key is the path to the private key for the HTTP server. May not be set.
	HTTPx509Key string `filepath:"true" yaml:"httpx509Key" json:"httpx509Key" env:"RQLITE_HTTP_X509_KEY"`

	// HTTPVerifyClient indicates whether the HTTP server should verify client certificates.
	HTTPVerifyClient bool `yaml:"httpVerifyClient" json:"httpVerifyClient" env:"RQLITE_HTTP_VERIFY_CLIENT"`

	// NodeX509CACert is the path to the CA certificate file for when this node verifies
	// other certificates for any inter-node communications. May not be set.
	NodeX509CACert string `filepath:"true" yaml:"nodeX509CACert" json:"nodeX509CACert" env:"RQLITE_NODE_X509_CA_CERT"`

	// NodeX509Cert is the path to the X509 cert for the Raft server. May not be set.
	NodeX509Cert string `filepath:"true" yaml:"nodeX509Cert" json:"nodeX509Cert" env:"RQLITE_NODE_X509_CERT"`

	// NodeX509Key is the path to the X509 key for the Raft server. May not be set.
	NodeX509Key string `filepath:"true" yaml:"nodeX509Key" json:"nodeX509Key" env:"RQLITE_NODE_X509_KEY"`

	// NoNodeVerify disables checking other nodes' Node X509 certs for validity.
	NoNodeVerify bool `yaml:"noNodeVerify" json:"noNodeVerify" env:"RQLITE_NO_NODE_VERIFY"`

	// NodeVerifyClient enable mutual TLS for node-to-node communication.
	NodeVerifyClient bool `yaml:"nodeVerifyClient" json:"nodeVerifyClient" env:"RQLITE_NODE_VERIFY_CLIENT"`

	// NodeVerifyServerName is the hostname to verify on the certificates returned by nodes.
	// If NoNodeVerify is true this field is ignored.
	NodeVerifyServerName string `yaml:"nodeVerifyServerName" json:"nodeVerifyServerName" env:"RQLITE_NODE_VERIFY_SERVER_NAME"`

	// NodeID is the Raft ID for the node.
	NodeID string `yaml:"nodeId" json:"nodeId" env:"RQLITE_NODE_ID"`

	// RaftAddr is the bind network address for the Raft server.
	RaftAddr string `yaml:"raftAddr" json:"raftAddr" env:"RQLITE_RAFT_ADDR"`

	// RaftAdv is the advertised Raft server address.
	RaftAdv string `yaml:"raftAdv" json:"raftAdv" env:"RQLITE_RAFT_ADV"`

	// JoinAddrs is the list of Raft addresses to use for a join attempt.
	JoinAddrs string `yaml:"joinAddrs" json:"joinAddrs" env:"RQLITE_JOIN_ADDRS"`

	// JoinAttempts is the number of times a node should attempt to join using a
	// given address.
	JoinAttempts int `yaml:"joinAttempts" json:"joinAttempts" env:"RQLITE_JOIN_ATTEMPTS"`

	// JoinInterval is the time between retrying failed join operations.
	JoinInterval time.Duration `yaml:"joinInterval" json:"joinInterval" env:"RQLITE_JOIN_INTERVAL"`

	// JoinAs sets the user join attempts should be performed as. May not be set.
	JoinAs string `yaml:"joinAs" json:"joinAs" env:"RQLITE_JOIN_AS"`

	// BootstrapExpect is the minimum number of nodes required for a bootstrap.
	BootstrapExpect int `yaml:"bootstrapExpect" json:"bootstrapExpect" env:"RQLITE_BOOTSTRAP_EXPECT"`

	// BootstrapExpectTimeout is the maximum time a bootstrap operation can take.
	BootstrapExpectTimeout time.Duration `yaml:"bootstrapExpectTimeout" json:"bootstrapExpectTimeout" env:"RQLITE_BOOTSTRAP_EXPECT_TIMEOUT"`

	// DiscoMode sets the discovery mode. May not be set.
	DiscoMode string `yaml:"discoMode" json:"discoMode" env:"RQLITE_DISCO_MODE"`

	// DiscoKey sets the discovery prefix key.
	DiscoKey string `yaml:"discoKey" json:"discoKey" env:"RQLITE_DISCO_KEY"`

	// DiscoConfig sets the path to any discovery configuration file. May not be set.
	DiscoConfig string `yaml:"discoConfig" json:"discoConfig" env:"RQLITE_DISCO_CONFIG"`

	// OnDiskPath sets the path to the SQLite file. May not be set.
	OnDiskPath string `yaml:"onDiskPath" json:"onDiskPath" env:"RQLITE_ON_DISK_PATH"`

	// FKConstraints enables SQLite foreign key constraints.
	FKConstraints bool `yaml:"fkConstraints" json:"fkConstraints" env:"RQLITE_FK_CONSTRAINTS"`

	// AutoVacInterval sets the automatic VACUUM interval. Use 0s to disable.
	AutoVacInterval time.Duration `yaml:"autoVacInterval" json:"autoVacInterval" env:"RQLITE_AUTO_VAC_INTERVAL"`

	// RaftLogLevel sets the minimum logging level for the Raft subsystem.
	RaftLogLevel string `yaml:"raftLogLevel" json:"raftLogLevel" env:"RQLITE_RAFT_LOG_LEVEL"`

	// RaftNonVoter controls whether this node is a voting, read-only node.
	RaftNonVoter bool `yaml:"raftNonVoter" json:"raftNonVoter" env:"RQLITE_RAFT_NON_VOTER"`

	// RaftSnapThreshold is the number of outstanding log entries that trigger snapshot.
	RaftSnapThreshold uint64 `yaml:"raftSnapThreshold" json:"raftSnapThreshold" env:"RQLITE_RAFT_SNAP_THRESHOLD"`

	// RaftSnapThreshold is the size of a SQLite WAL file which will trigger a snapshot.
	RaftSnapThresholdWALSize uint64 `yaml:"raftSnapThresholdWALSize" json:"raftSnapThresholdWALSize" env:"RQLITE_"`

	// RaftSnapInterval sets the threshold check interval.
	RaftSnapInterval time.Duration `yaml:"raftSnapInterval" json:"raftSnapInterval" env:"RQLITE_RAFT_SNAP_INTERVAL"`

	// RaftLeaderLeaseTimeout sets the leader lease timeout.
	RaftLeaderLeaseTimeout time.Duration `yaml:"raftLeaderLeaseTimeout" json:"raftLeaderLeaseTimeout" env:"RQLITE_RAFT_LEADER_LEASE_TIMEOUT"`

	// RaftHeartbeatTimeout specifies the time in follower state without contact
	// from a Leader before the node attempts an election.
	RaftHeartbeatTimeout time.Duration `yaml:"raftHeartbeatTimeout" json:"raftHeartbeatTimeout" env:"RQLITE_HEARTBEAT_TIMEOUT"`

	// RaftElectionTimeout specifies the time in candidate state without contact
	// from a Leader before the node attempts an election.
	RaftElectionTimeout time.Duration `yaml:"raftElectionTimeout" json:"raftElectionTimeout" env:"RQLITE_RAFT_ELECTION_TIMEOUT"`

	// RaftApplyTimeout sets the Log-apply timeout.
	RaftApplyTimeout time.Duration `yaml:"raftApplyTimeout" json:"raftApplyTimeout" env:"RQLITE_RAFT_APPLY_TIMEOUT"`

	// RaftShutdownOnRemove sets whether Raft should be shutdown if the node is removed
	RaftShutdownOnRemove bool `yaml:"raftShutdownOnRemove" json:"raftShutdownOnRemove" env:"RQLITE_RAFT_SHUTDOWN_ON_REMOVE"`

	// RaftClusterRemoveOnShutdown sets whether the node should remove itself from the cluster on shutdown
	RaftClusterRemoveOnShutdown bool `yaml:"raftClusterRemoveOnShutdown" json:"raftClusterRemoveOnShutdown" env:"RQLITE_RAFT_CLUSTER_REMOVE_ON_SHUTDOWN"`

	// RaftStepdownOnShutdown sets whether Leadership should be relinquished on shutdown
	RaftStepdownOnShutdown bool `yaml:"raftStepdownOnShutdown" json:"raftStepdownOnShutdown" env:"RQLITE_RAFT_STEPDOWN_ON_SHUTDOWN"`

	// RaftReapNodeTimeout sets the duration after which a non-reachable voting node is
	// reaped i.e. removed from the cluster.
	RaftReapNodeTimeout time.Duration `yaml:"raftReapNodeTimeout" json:"raftReapNodeTimeout" env:"RQLITE_RAFT_REAP_NODE_TIMEOUT"`

	// RaftReapReadOnlyNodeTimeout sets the duration after which a non-reachable non-voting node is
	// reaped i.e. removed from the cluster.
	RaftReapReadOnlyNodeTimeout time.Duration `yaml:"raftReapReadOnlyNodeTimeout" json:"raftReapReadOnlyNodeTimeout" env:"RQLITE_RAFT_REAP_READ_ONLY_NODE_TIMEOUT"`

	// ClusterConnectTimeout sets the timeout when initially connecting to another node in
	// the cluster, for non-Raft communications.
	ClusterConnectTimeout time.Duration `yaml:"clusterConnectTimeout" json:"clusterConnectTimeout" env:"RQLITE_CLUSTER_CONNECT_TIMEOUT"`

	// WriteQueueCap is the default capacity of Execute queues
	WriteQueueCap int `yaml:"writeQueueCap" json:"writeQueueCap" env:"RQLITE_WRITE_QUEUE_CAP"`

	// WriteQueueBatchSz is the default batch size for Execute queues
	WriteQueueBatchSz int `yaml:"writeQueueBatchSz" json:"writeQueueBatchSz" env:"RQLITE_WRITE_QUEUE_BATCH_SIZE"`

	// WriteQueueTimeout is the default time after which any data will be sent on
	// Execute queues, if a batch size has not been reached.
	WriteQueueTimeout time.Duration `yaml:"writeQueueTimeout" json:"writeQueueTimeout" env:"RQLITE_WRITE_QUEUE_TIMEOUT"`

	// WriteQueueTx controls whether writes from the queue are done within a transaction.
	WriteQueueTx bool `yaml:"writeQueueTx" json:"writeQueueTx" env:"RQLITE_WRITE_QUEUE_TX"`

	// CPUProfile enables CPU profiling.
	CPUProfile string `yaml:"cpuProfile" json:"cpuProfile" env:"RQLITE_CPU_PROFILE"`

	// MemProfile enables memory profiling.
	MemProfile string `yaml:"memProfile" json:"memProfile" env:"RQLITE_MEM_PROFILE"`
}

// Validate checks the configuration for internal consistency, and activates
// important rqlite policies. It must be called at least once on a Config
// object before the Config object is used. It is OK to call more than
// once.
func (c *RqLite) Validate() error {
	dataPath, err := filepath.Abs(c.DataPath)
	if err != nil {
		return fmt.Errorf("failed to determine absolute data path: %s", err.Error())
	}
	c.DataPath = dataPath

	err = c.CheckFilePaths()
	if err != nil {
		return err
	}

	if !bothUnsetSet(c.HTTPx509Cert, c.HTTPx509Key) {
		return fmt.Errorf("either both HTTPx509Cert and HTTPx509Key must be set, or neither")
	}
	if !bothUnsetSet(c.NodeX509Cert, c.NodeX509Key) {
		return fmt.Errorf("either both HTTPx509Cert and HTTPx509Key must be set, or neither")

	}

	if c.RaftAddr == c.HTTPAddr {
		return errors.New("HTTP and Raft addresses must differ")
	}

	// Enforce policies regarding addresses
	if c.RaftAdv == "" {
		c.RaftAdv = c.RaftAddr
	}
	if c.HTTPAdv == "" {
		c.HTTPAdv = c.HTTPAddr
	}

	// Node ID policy
	if c.NodeID == "" {
		c.NodeID = c.RaftAdv
	}

	// Perform some address validity checks.
	if strings.HasPrefix(strings.ToLower(c.HTTPAddr), "http") ||
		strings.HasPrefix(strings.ToLower(c.HTTPAdv), "http") {
		return errors.New("HTTP options should not include protocol (http:// or https://)")
	}
	if _, _, err := net.SplitHostPort(c.HTTPAddr); err != nil {
		return errors.New("HTTP bind address not valid")
	}

	hadv, _, err := net.SplitHostPort(c.HTTPAdv)
	if err != nil {
		return errors.New("HTTP advertised HTTP address not valid")
	}
	if addr := net.ParseIP(hadv); addr != nil && addr.IsUnspecified() {
		return fmt.Errorf("advertised HTTP address is not routable (%s), specify it via HTTPAdv or HTTPAddr",
			hadv)
	}

	if _, rp, err := net.SplitHostPort(c.RaftAddr); err != nil {
		return errors.New("raft bind address not valid")
	} else if _, err := strconv.Atoi(rp); err != nil {
		return errors.New("raft bind port not valid")
	}

	radv, rp, err := net.SplitHostPort(c.RaftAdv)
	if err != nil {
		return errors.New("raft advertised address not valid")
	}
	if addr := net.ParseIP(radv); addr != nil && addr.IsUnspecified() {
		return fmt.Errorf("advertised Raft address is not routable (%s), specify it via RaftAddr or RaftAdv",
			radv)
	}
	if _, err := strconv.Atoi(rp); err != nil {
		return errors.New("raft advertised port is not valid")
	}

	if c.RaftAdv == c.HTTPAdv {
		return errors.New("advertised HTTP and Raft addresses must differ")
	}

	// Enforce bootstrapping policies
	if c.BootstrapExpect > 0 && c.RaftNonVoter {
		return errors.New("bootstrapping only applicable to voting nodes")
	}

	// Join parameters OK?
	if c.JoinAddrs != "" {
		addrs := strings.Split(c.JoinAddrs, ",")
		for i := range addrs {
			if _, _, err := net.SplitHostPort(addrs[i]); err != nil {
				return fmt.Errorf("%s is an invalid join address", addrs[i])
			}

			if c.BootstrapExpect == 0 {
				if addrs[i] == c.RaftAdv || addrs[i] == c.RaftAddr {
					return errors.New("node cannot join with itself unless bootstrapping")
				}
				if c.AutoRestoreFile != "" {
					return errors.New("auto-restoring cannot be used when joining a cluster")
				}
			}
		}

		if c.DiscoMode != "" {
			return errors.New("disco mode cannot be used when also explicitly joining a cluster")
		}
	}

	// Valid disco mode?
	switch c.DiscoMode {
	case "":
	case DiscoModeEtcdKV, DiscoModeConsulKV:
		if c.BootstrapExpect > 0 {
			return fmt.Errorf("bootstrapping not applicable when using %s", c.DiscoMode)
		}
	case DiscoModeDNS, DiscoModeDNSSRV:
		if c.BootstrapExpect == 0 && !c.RaftNonVoter {
			return fmt.Errorf("bootstrap-expect value required when using %s with a voting node", c.DiscoMode)
		}
	default:
		return fmt.Errorf("disco mode must be one of %s, %s, %s, or %s",
			DiscoModeConsulKV, DiscoModeEtcdKV, DiscoModeDNS, DiscoModeDNSSRV)
	}

	return nil
}

// JoinAddresses returns the join addresses set at the command line. Returns nil
// if no join addresses were set.
func (c *RqLite) JoinAddresses() []string {
	if c.JoinAddrs == "" {
		return nil
	}
	return strings.Split(c.JoinAddrs, ",")
}

// HTTPURL returns the fully-formed, advertised HTTP API address for this config, including
// protocol, host and port.
func (c *RqLite) HTTPURL() string {
	apiProto := "http"
	if c.HTTPx509Cert != "" {
		apiProto = "https"
	}
	// return fmt.Sprintf("%s://%s", apiProto, c.HTTPAdv)
	return fmt.Sprintf("%s://%s", apiProto, "")
}

// RaftPort returns the port on which the Raft system is listening. Validate must
// have been called before calling this method.
func (c *RqLite) RaftPort() int {
	_, port, err := net.SplitHostPort(c.RaftAddr)
	if err != nil {
		panic("RaftAddr not valid")
	}
	p, err := strconv.Atoi(port)
	if err != nil {
		panic("RaftAddr port not valid")
	}
	return p
}

// DiscoConfigReader returns a ReadCloser providing access to the Disco config.
// The caller must call close on the ReadCloser when finished with it. If no
// config was supplied, it returns nil.
func (c *RqLite) DiscoConfigReader() io.ReadCloser {
	var rc io.ReadCloser
	if c.DiscoConfig == "" {
		return nil
	}

	// Open config file. If opening fails, assume string is the literal config.
	cfgFile, err := os.Open(c.DiscoConfig)
	if err != nil {
		rc = io.NopCloser(bytes.NewReader([]byte(c.DiscoConfig)))
	} else {
		rc = cfgFile
	}
	return rc
}

// CheckFilePaths checks that all file paths in the config exist.
// Empty filepaths are ignored.
func (c *RqLite) CheckFilePaths() error {
	v := reflect.ValueOf(c).Elem()

	// Iterate through the fields of the struct
	for i := 0; i < v.NumField(); i++ {
		field := v.Type().Field(i)
		fieldValue := v.Field(i)

		if fieldValue.Kind() != reflect.String {
			continue
		}

		if tagValue, ok := field.Tag.Lookup("filepath"); ok && tagValue == "true" {
			filePath := fieldValue.String()
			if filePath == "" {
				continue
			}
			_, err := os.Stat(filePath)
			if os.IsNotExist(err) {
				return fmt.Errorf("%s does not exist", filePath)
			}
		}
	}
	return nil
}

// bothUnsetSet returns true if both a and b are unset, or both are set.
func bothUnsetSet(a, b string) bool {
	return (a == "" && b == "") || (a != "" && b != "")
}
