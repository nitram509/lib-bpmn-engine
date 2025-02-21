package cluster

import (
	"strings"
	"time"

	"github.com/pbinitiative/zenbpm/internal/config"
)

func GetDefaultConfig(nodeId string, raftAddr string, httpAddr string, dataPath string, joinAddresses []string) config.RqLite {
	return config.RqLite{
		DataPath:                    dataPath,
		AuthFile:                    "",
		AutoBackupFile:              "",
		AutoRestoreFile:             "",
		HTTPAddr:                    httpAddr,
		HTTPAdv:                     httpAddr,
		HTTPx509CACert:              "",
		HTTPx509Cert:                "",
		HTTPx509Key:                 "",
		HTTPVerifyClient:            false,
		NodeX509CACert:              "",
		NodeX509Cert:                "",
		NodeX509Key:                 "",
		NoNodeVerify:                true,
		NodeVerifyClient:            false,
		NodeVerifyServerName:        "",
		NodeID:                      nodeId,
		RaftAddr:                    raftAddr,
		RaftAdv:                     raftAddr,
		JoinAddrs:                   strings.Join(joinAddresses, ","),
		JoinAttempts:                5,
		JoinInterval:                3 * time.Second,
		JoinAs:                      "",
		BootstrapExpect:             1,
		BootstrapExpectTimeout:      120 * time.Second,
		DiscoMode:                   config.DiscoModeNone,
		DiscoKey:                    "",
		DiscoConfig:                 "",
		OnDiskPath:                  "",
		FKConstraints:               false,
		AutoVacInterval:             12 * time.Hour,
		RaftLogLevel:                "WARN",
		RaftNonVoter:                false,
		RaftSnapThreshold:           8192,
		RaftSnapThresholdWALSize:    4 * 1024 * 1024,
		RaftSnapInterval:            10 * time.Second,
		RaftLeaderLeaseTimeout:      0,
		RaftHeartbeatTimeout:        1 * time.Second,
		RaftElectionTimeout:         1 * time.Second,
		RaftApplyTimeout:            10 * time.Second,
		RaftShutdownOnRemove:        false,
		RaftClusterRemoveOnShutdown: true,
		RaftStepdownOnShutdown:      true,
		RaftReapNodeTimeout:         0,
		RaftReapReadOnlyNodeTimeout: 0,
		ClusterConnectTimeout:       30 * time.Second,
		WriteQueueCap:               1024,
		WriteQueueBatchSz:           128,
		WriteQueueTimeout:           50 * time.Millisecond,
		WriteQueueTx:                false,
		CPUProfile:                  "",
		MemProfile:                  "",
	}
}
