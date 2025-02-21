package cluster

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/pbinitiative/zenbpm/internal/config"
	"github.com/pbinitiative/zenbpm/internal/log"
	"github.com/rqlite/rqlite-disco-clients/consul"
	"github.com/rqlite/rqlite-disco-clients/dns"
	"github.com/rqlite/rqlite-disco-clients/dnssrv"
	etcd "github.com/rqlite/rqlite-disco-clients/etcd"
	"github.com/rqlite/rqlite/v8/auth"
	"github.com/rqlite/rqlite/v8/auto/backup"
	"github.com/rqlite/rqlite/v8/auto/restore"
	"github.com/rqlite/rqlite/v8/aws"
	"github.com/rqlite/rqlite/v8/cluster"
	"github.com/rqlite/rqlite/v8/command/proto"
	"github.com/rqlite/rqlite/v8/disco"
	httpd "github.com/rqlite/rqlite/v8/http"
	"github.com/rqlite/rqlite/v8/rtls"
	"github.com/rqlite/rqlite/v8/store"
	"github.com/rqlite/rqlite/v8/tcp"
)

// ZenPartitionNode is part of the rqlite raft cluster (partition of the main cluster)
type ZenPartitionNode struct {
	config          *config.RqLite
	store           *store.Store
	muxListener     net.Listener
	credentialStore *auth.CredentialsStore
	clusterClient   *cluster.Client
	clusterService  *cluster.Service
	statusMu        sync.Mutex
	statuses        map[string]httpd.StatusReporter
}

func (c *ZenPartitionNode) RegisterStatus(key string, stat httpd.StatusReporter) error {
	c.statusMu.Lock()
	defer c.statusMu.Unlock()

	if _, ok := c.statuses[key]; ok {
		return fmt.Errorf("status already registered with key %s", key)
	}
	c.statuses[key] = stat

	return nil
}

func (c *ZenPartitionNode) IsLeader(ctx context.Context) bool {
	return c.store.IsLeader()
}

// Execute an SQL statement on rqlite partition
func (c *ZenPartitionNode) Execute(ctx context.Context, req *proto.ExecuteRequest) ([]*proto.ExecuteQueryResponse, error) {
	return c.store.Execute(req)
}

// Run an SQL query on rqlite partition
func (c *ZenPartitionNode) Query(ctx context.Context, req *proto.QueryRequest) ([]*proto.QueryRows, error) {
	return c.store.Query(req)
}

func StartZenPartitionNode(mainCtx context.Context, cfg *config.RqLite) (*ZenPartitionNode, error) {
	// struct that will hold our partition cluster
	zenPartitionNode := ZenPartitionNode{
		config:   cfg,
		statuses: map[string]httpd.StatusReporter{},
	}

	// Create internode network mux and configure.
	muxLn, err := net.Listen("tcp", cfg.RaftAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on %s: %w", cfg.RaftAddr, err)
	}
	zenPartitionNode.muxListener = muxLn
	mux, err := startNodeMux(cfg, muxLn)
	if err != nil {
		return nil, fmt.Errorf("failed to start node mux: %w", err)
	}

	// Raft internode layer
	raftLn := mux.Listen(cluster.MuxRaftHeader)
	log.Info("Raft TCP mux Listener registered with byte header %d", cluster.MuxRaftHeader)
	raftDialer, err := cluster.CreateRaftDialer(cfg.NodeX509Cert, cfg.NodeX509Key, cfg.NodeX509CACert,
		cfg.NodeVerifyServerName, cfg.NoNodeVerify)
	if err != nil {
		return nil, fmt.Errorf("failed to create Raft dialer: %w", err)
	}
	raftTn := tcp.NewLayer(raftLn, raftDialer)

	// Create the store.
	str, err := createStore(cfg, raftTn)
	if err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}
	zenPartitionNode.store = str

	// Install the auto-restore data, if necessary.
	if cfg.AutoRestoreFile != "" {
		hd, err := store.HasData(str.Path())
		if err != nil {
			return nil, fmt.Errorf("failed to check for existing data: %w", err)
		}
		if hd {
			log.Info("auto-restore requested, but data already exists in %s, skipping", str.Path())
		} else {
			log.Info("auto-restore requested, initiating download")
			start := time.Now()
			path, errOK, err := restore.DownloadFile(mainCtx, cfg.AutoRestoreFile)
			if err != nil {
				var b strings.Builder
				b.WriteString(fmt.Sprintf("failed to download auto-restore file: %s", err.Error()))
				if errOK {
					b.WriteString(", continuing with node startup anyway")
					log.Info(b.String())
				} else {
					return nil, fmt.Errorf(b.String())
				}
			} else {
				log.Info("auto-restore file downloaded in %s", time.Since(start))
				if err := str.SetRestorePath(path); err != nil {
					return nil, fmt.Errorf("failed to preload auto-restore data: %w", err)
				}
			}
		}
	}

	// Get any credential store.
	credStr, err := credentialStore(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to get credential store: %w", err)
	}
	zenPartitionNode.credentialStore = credStr

	// Create cluster service now, so nodes will be able to learn information about each other.
	clstrServ, err := clusterService(cfg, mux.Listen(cluster.MuxClusterHeader), str, str, credStr)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster service: %w", err)
	}
	zenPartitionNode.clusterService = clstrServ
	log.Info("cluster TCP mux Listener registered with byte header %d", cluster.MuxClusterHeader)

	// Create the HTTP service.
	//
	// We want to start the HTTP server as soon as possible, so the node is responsive and external
	// systems can see that it's running. We still have to open the Store though, so the node won't
	// be able to do much until that happens however.
	clstrClient, err := createClusterClient(cfg, clstrServ)
	if err != nil {
		return nil, fmt.Errorf("failed to create cluster client: %w", err)
	}
	zenPartitionNode.clusterClient = clstrClient
	// httpServ, err := startHTTPService(cfg, str, clstrClient, credStr)
	// if err != nil {
	// 	log.Fatalf("failed to start HTTP server: %s", err.Error())
	// }
	// partitionCluster.service = httpServ

	// Now, open store. How long this takes does depend on how much data is being stored by rqlite.
	if err := str.Open(); err != nil {
		return nil, fmt.Errorf("failed to open store: %w", err)
	}

	// Register remaining status providers.
	zenPartitionNode.RegisterStatus("cluster", clstrServ)
	zenPartitionNode.RegisterStatus("network", tcp.NetworkReporter{})

	// Create the cluster!
	nodes, err := str.Nodes()
	if err != nil {
		return nil, fmt.Errorf("failed to get nodes %w", err)
	}
	if err := createCluster(mainCtx, cfg, len(nodes) > 0, &zenPartitionNode); err != nil {
		return nil, fmt.Errorf("clustering failure: %w", err)
	}

	// Start any requested auto-backups
	backupSrv, err := startAutoBackups(mainCtx, cfg, str)
	if err != nil {
		return nil, fmt.Errorf("failed to start auto-backups: %w", err)
	}
	if backupSrv != nil {
		zenPartitionNode.RegisterStatus("auto_backups", backupSrv)
	}
	return &zenPartitionNode, nil
}

func (partitionCluster *ZenPartitionNode) Stats() (map[string]interface{}, error) {
	return partitionCluster.clusterClient.Stats()
}

func (partitionCluster *ZenPartitionNode) Stop() error {
	// Stop the HTTP server first, so clients get notification as soon as
	// possible that the node is going away.

	if partitionCluster.config.RaftClusterRemoveOnShutdown {
		remover := cluster.NewRemover(partitionCluster.clusterClient, 5*time.Second, partitionCluster.store)
		remover.SetCredentials(cluster.CredentialsFor(partitionCluster.credentialStore, partitionCluster.config.JoinAs))
		log.Info("initiating removal of this node from cluster before shutdown")
		if err := remover.Do(partitionCluster.config.NodeID, true); err != nil {
			return fmt.Errorf("failed to remove this node from cluster before shutdown: %w", err)
		}
		log.Info("removed this node successfully from cluster before shutdown")
	}

	if partitionCluster.config.RaftStepdownOnShutdown {
		if partitionCluster.store.IsLeader() {
			// Don't log a confusing message if (probably) not Leader
			log.Info("stepping down as Leader before shutdown")
		}
		// Perform a stepdown, ignore any errors.
		partitionCluster.store.Stepdown(true)
	}

	if err := partitionCluster.store.Close(true); err != nil {
		log.Info("failed to close store: %s", err.Error())
	}
	partitionCluster.clusterService.Close()
	partitionCluster.muxListener.Close()
	log.Info("rqlite server stopped")
	return nil
}

func startAutoBackups(ctx context.Context, cfg *config.RqLite, str *store.Store) (*backup.Uploader, error) {
	if cfg.AutoBackupFile == "" {
		return nil, nil
	}

	b, err := backup.ReadConfigFile(cfg.AutoBackupFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read auto-backup file: %s", err.Error())
	}

	uCfg, s3cfg, err := backup.Unmarshal(b)
	if err != nil {
		return nil, fmt.Errorf("failed to parse auto-backup file: %s", err.Error())
	}
	provider := store.NewProvider(str, uCfg.Vacuum, !uCfg.NoCompress)
	sc, err := aws.NewS3Client(s3cfg.Endpoint, s3cfg.Region, s3cfg.AccessKeyID, s3cfg.SecretAccessKey,
		s3cfg.Bucket, s3cfg.Path, &aws.S3ClientOpts{
			ForcePathStyle: s3cfg.ForcePathStyle,
		})
	if err != nil {
		return nil, fmt.Errorf("failed to create aws S3 client: %s", err.Error())
	}
	u := backup.NewUploader(sc, provider, time.Duration(uCfg.Interval))
	u.Start(ctx, str.IsLeader)
	return u, nil
}

func createStore(cfg *config.RqLite, ln *tcp.Layer) (*store.Store, error) {
	dbConf := store.NewDBConfig()
	dbConf.OnDiskPath = cfg.OnDiskPath
	dbConf.FKConstraints = cfg.FKConstraints

	str := store.New(ln, &store.Config{
		DBConf: dbConf,
		Dir:    cfg.DataPath,
		ID:     cfg.NodeID,
		Logger: hclog.Default().
			Named("store").
			StandardLogger(&hclog.StandardLoggerOptions{
				ForceLevel: hclog.Default().GetLevel(),
			}),
	})

	// Set optional parameters on store.
	str.RaftLogLevel = cfg.RaftLogLevel
	str.ShutdownOnRemove = cfg.RaftShutdownOnRemove
	str.SnapshotThreshold = cfg.RaftSnapThreshold
	str.SnapshotThresholdWALSize = cfg.RaftSnapThresholdWALSize
	str.SnapshotInterval = cfg.RaftSnapInterval
	str.LeaderLeaseTimeout = cfg.RaftLeaderLeaseTimeout
	str.HeartbeatTimeout = cfg.RaftHeartbeatTimeout
	str.ElectionTimeout = cfg.RaftElectionTimeout
	str.ApplyTimeout = cfg.RaftApplyTimeout
	str.BootstrapExpect = cfg.BootstrapExpect
	str.ReapTimeout = cfg.RaftReapNodeTimeout
	str.ReapReadOnlyTimeout = cfg.RaftReapReadOnlyNodeTimeout
	str.AutoVacInterval = cfg.AutoVacInterval

	if store.IsNewNode(cfg.DataPath) {
		log.Info("no preexisting node state detected in %s, node may be bootstrapping", cfg.DataPath)
	} else {
		log.Info("preexisting node state detected in %s", cfg.DataPath)
	}

	return str, nil
}

func createDiscoService(cfg *config.RqLite, str *store.Store) (*disco.Service, error) {
	var c disco.Client
	var err error

	rc := cfg.DiscoConfigReader()
	defer func() {
		if rc != nil {
			rc.Close()
		}
	}()
	if cfg.DiscoMode == config.DiscoModeConsulKV {
		var consulCfg *consul.Config
		consulCfg, err = consul.NewConfigFromReader(rc)
		if err != nil {
			return nil, fmt.Errorf("create Consul config: %s", err.Error())
		}

		c, err = consul.New(cfg.DiscoKey, consulCfg)
		if err != nil {
			return nil, fmt.Errorf("create Consul client: %s", err.Error())
		}
	} else if cfg.DiscoMode == config.DiscoModeEtcdKV {
		var etcdCfg *etcd.Config
		etcdCfg, err = etcd.NewConfigFromReader(rc)
		if err != nil {
			return nil, fmt.Errorf("create etcd config: %s", err.Error())
		}

		c, err = etcd.New(cfg.DiscoKey, etcdCfg)
		if err != nil {
			return nil, fmt.Errorf("create etcd client: %s", err.Error())
		}
	} else {
		return nil, fmt.Errorf("invalid disco service: %s", cfg.DiscoMode)
	}
	return disco.NewService(c, str, disco.VoterSuffrage(!cfg.RaftNonVoter)), nil
}

// TODO: Remove once the cluster is set up. We do not want to expose rqLite cluster through http interface
// func startHTTPService(cfg *Config, str *store.Store, cltr *cluster.Client, credStr *auth.CredentialsStore) (*Service, error) {
// 	// Create HTTP server and load authentication information.
// 	s := New(cfg.HTTPAddr, str, cltr, credStr)
//
// 	s.CACertFile = cfg.HTTPx509CACert
// 	s.CertFile = cfg.HTTPx509Cert
// 	s.KeyFile = cfg.HTTPx509Key
// 	s.ClientVerify = cfg.HTTPVerifyClient
// 	s.DefaultQueueCap = cfg.WriteQueueCap
// 	s.DefaultQueueBatchSz = cfg.WriteQueueBatchSz
// 	s.DefaultQueueTimeout = cfg.WriteQueueTimeout
// 	s.DefaultQueueTx = cfg.WriteQueueTx
// 	s.AllowOrigin = cfg.HTTPAllowOrigin
// 	s.BuildInfo = map[string]interface{}{
// 		"commit":     cmd.Commit,
// 		"branch":     cmd.Branch,
// 		"version":    cmd.Version,
// 		"compiler":   runtime.Compiler,
// 		"build_time": cmd.Buildtime,
// 	}
// 	return s, s.Start()
// }

// startNodeMux starts the TCP mux on the given listener, which should be already
// bound to the relevant interface.
func startNodeMux(cfg *config.RqLite, ln net.Listener) (*tcp.Mux, error) {
	var err error
	adv := tcp.NameAddress{
		Address: cfg.RaftAdv,
	}

	var mux *tcp.Mux
	if cfg.NodeX509Cert != "" {
		var b strings.Builder
		b.WriteString(fmt.Sprintf("enabling node-to-node encryption with cert: %s, key: %s",
			cfg.NodeX509Cert, cfg.NodeX509Key))
		if cfg.NodeX509CACert != "" {
			b.WriteString(fmt.Sprintf(", CA cert %s", cfg.NodeX509CACert))
		}
		if cfg.NodeVerifyClient {
			b.WriteString(", mutual TLS enabled")
		} else {
			b.WriteString(", mutual TLS disabled")
		}
		log.Info(b.String())
		mux, err = tcp.NewTLSMux(ln, adv, cfg.NodeX509Cert, cfg.NodeX509Key, cfg.NodeX509CACert,
			cfg.NoNodeVerify, cfg.NodeVerifyClient)
	} else {
		mux, err = tcp.NewMux(ln, adv)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create node-to-node mux: %s", err.Error())
	}
	go mux.Serve()
	return mux, nil
}

func credentialStore(cfg *config.RqLite) (*auth.CredentialsStore, error) {
	if cfg.AuthFile == "" {
		return nil, nil
	}
	return auth.NewCredentialsStoreFromFile(cfg.AuthFile)
}

func clusterService(cfg *config.RqLite, ln net.Listener, db cluster.Database, mgr cluster.Manager, credStr *auth.CredentialsStore) (*cluster.Service, error) {
	c := cluster.New(ln, db, mgr, credStr)
	c.SetAPIAddr(cfg.HTTPAdv)
	c.EnableHTTPS(cfg.HTTPx509Cert != "" && cfg.HTTPx509Key != "") // Conditions met for an HTTPS API
	if err := c.Open(); err != nil {
		return nil, err
	}
	return c, nil
}

func createClusterClient(cfg *config.RqLite, clstr *cluster.Service) (*cluster.Client, error) {
	var dialerTLSConfig *tls.Config
	var err error
	if cfg.NodeX509Cert != "" || cfg.NodeX509CACert != "" {
		dialerTLSConfig, err = rtls.CreateClientConfig(cfg.NodeX509Cert, cfg.NodeX509Key,
			cfg.NodeX509CACert, cfg.NodeVerifyServerName, cfg.NoNodeVerify)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config for cluster dialer: %s", err.Error())
		}
	}
	clstrDialer := tcp.NewDialer(cluster.MuxClusterHeader, dialerTLSConfig)
	clstrClient := cluster.NewClient(clstrDialer, cfg.ClusterConnectTimeout)
	if err := clstrClient.SetLocal(cfg.RaftAdv, clstr); err != nil {
		return nil, fmt.Errorf("failed to set cluster client local parameters: %s", err.Error())
	}
	return clstrClient, nil
}

func createCluster(ctx context.Context, cfg *config.RqLite, hasPeers bool, partitionCluster *ZenPartitionNode) error {
	joins := cfg.JoinAddresses()
	if err := networkCheckJoinAddrs(joins); err != nil {
		return err
	}
	if joins == nil && cfg.DiscoMode == "" && !hasPeers {
		if cfg.RaftNonVoter {
			return fmt.Errorf("cannot create a new non-voting node without joining it to an existing cluster")
		}

		// Brand new node, told to bootstrap itself. So do it.
		log.Info("bootstrapping single new node")
		if err := partitionCluster.store.Bootstrap(store.NewServer(partitionCluster.store.ID(), cfg.RaftAdv, true)); err != nil {
			return fmt.Errorf("failed to bootstrap single new node: %s", err.Error())
		}
		return nil
	}

	// Prepare definition of being part of a cluster.
	bootDoneFn := func() bool {
		leader, _ := partitionCluster.store.LeaderAddr()
		return leader != ""
	}
	clusterSuf := cluster.VoterSuffrage(!cfg.RaftNonVoter)

	joiner := cluster.NewJoiner(partitionCluster.clusterClient, cfg.JoinAttempts, cfg.JoinInterval)
	joiner.SetCredentials(cluster.CredentialsFor(partitionCluster.credentialStore, cfg.JoinAs))
	if joins != nil && cfg.BootstrapExpect == 0 {
		// Explicit join operation requested, so do it.
		j, err := joiner.Do(ctx, joins, partitionCluster.store.ID(), cfg.RaftAdv, clusterSuf)
		if err != nil {
			return fmt.Errorf("failed to join cluster: %s", err.Error())
		}
		log.Info("successfully joined cluster at", j)
		return nil
	}

	if joins != nil && cfg.BootstrapExpect > 0 {
		// Bootstrap with explicit join addresses requests.
		bs := cluster.NewBootstrapper(cluster.NewAddressProviderString(joins), partitionCluster.clusterClient)
		bs.SetCredentials(cluster.CredentialsFor(partitionCluster.credentialStore, cfg.JoinAs))
		return bs.Boot(ctx, partitionCluster.store.ID(), cfg.RaftAdv, clusterSuf, bootDoneFn, cfg.BootstrapExpectTimeout)
	}

	if cfg.DiscoMode == "" {
		// No more clustering techniques to try. Node will just sit, probably using
		// existing Raft state.
		return nil
	}

	// DNS-based discovery requested. It's OK to proceed with this even if this node
	// is already part of a cluster. Re-joining and re-notifying other nodes will be
	// ignored when the node is already part of the cluster.
	log.Info("discovery mode: %s", cfg.DiscoMode)
	switch cfg.DiscoMode {
	case config.DiscoModeDNS, config.DiscoModeDNSSRV:
		rc := cfg.DiscoConfigReader()
		defer func() {
			if rc != nil {
				rc.Close()
			}
		}()

		var provider interface {
			cluster.AddressProvider
			httpd.StatusReporter
		}
		if cfg.DiscoMode == config.DiscoModeDNS {
			dnsCfg, err := dns.NewConfigFromReader(rc)
			if err != nil {
				return fmt.Errorf("error reading DNS configuration: %s", err.Error())
			}
			provider = dns.NewWithPort(dnsCfg, cfg.RaftPort())

		} else {
			dnssrvCfg, err := dnssrv.NewConfigFromReader(rc)
			if err != nil {
				return fmt.Errorf("error reading DNS configuration: %s", err.Error())
			}
			provider = dnssrv.New(dnssrvCfg)
		}

		bs := cluster.NewBootstrapper(provider, partitionCluster.clusterClient)
		bs.SetCredentials(cluster.CredentialsFor(partitionCluster.credentialStore, cfg.JoinAs))
		partitionCluster.RegisterStatus("disco", provider)
		return bs.Boot(ctx, partitionCluster.store.ID(), cfg.RaftAdv, clusterSuf, bootDoneFn, cfg.BootstrapExpectTimeout)

	case config.DiscoModeEtcdKV, config.DiscoModeConsulKV:
		discoService, err := createDiscoService(cfg, partitionCluster.store)
		if err != nil {
			return fmt.Errorf("failed to start discovery service: %s", err.Error())
		}
		// Safe to start reporting before doing registration. If the node hasn't bootstrapped
		// yet, or isn't leader, reporting will just be a no-op until something changes.
		go discoService.StartReporting(cfg.NodeID, cfg.HTTPURL(), cfg.RaftAdv)
		partitionCluster.RegisterStatus("disco", discoService)

		if hasPeers {
			log.Info("preexisting node configuration detected, not registering with discovery service")
			return nil
		}
		log.Info("no preexisting nodes, registering with discovery service")

		leader, addr, err := discoService.Register(partitionCluster.store.ID(), cfg.HTTPURL(), cfg.RaftAdv)
		if err != nil {
			return fmt.Errorf("failed to register with discovery service: %s", err.Error())
		}
		if leader {
			log.Info("node registered as leader using discovery service")
			if err := partitionCluster.store.Bootstrap(store.NewServer(partitionCluster.store.ID(), partitionCluster.store.Addr(), true)); err != nil {
				return fmt.Errorf("failed to bootstrap single new node: %s", err.Error())
			}
		} else {
			for {
				log.Info("discovery service returned %s as join address", addr)
				if j, err := joiner.Do(ctx, []string{addr}, partitionCluster.store.ID(), cfg.RaftAdv, clusterSuf); err != nil {
					log.Info("failed to join cluster at %s: %s", addr, err.Error())

					time.Sleep(time.Second)
					_, addr, err = discoService.Register(partitionCluster.store.ID(), cfg.HTTPURL(), cfg.RaftAdv)
					if err != nil {
						log.Info("failed to get updated leader: %s", err.Error())
					}
					continue
				} else {
					log.Info("successfully joined cluster at", j)
					break
				}
			}
		}

	default:
		return fmt.Errorf("invalid disco mode %s", cfg.DiscoMode)
	}
	return nil
}

func networkCheckJoinAddrs(joinAddrs []string) error {
	if len(joinAddrs) > 0 {
		log.Info("checking that supplied join addresses don't serve HTTP(S)")
		if addr, ok := httpd.AnyServingHTTP(joinAddrs); ok {
			return fmt.Errorf("join address %s appears to be serving HTTP when it should be Raft", addr)
		}
	}
	return nil
}
