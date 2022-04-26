package app

import (
	"context"
	"minik8s/pkg/client"
	"minik8s/pkg/controller/replicaset"
	"minik8s/pkg/klog"
	"minik8s/pkg/messaging"
	"time"
)

func startReplicaSetController(ctx context.Context, controllerCtx ControllerContext) error {
	// todo initialize config file in global
	msgConfig := messaging.QConfig{
		User:          "111",
		Password:      "111",
		Host:          "127.0.0.1",
		Port:          "9000",
		MaxRetry:      5,
		RetryInterval: time.Duration(1000),
	}
	clientConfig := client.Config{
		Host: "127.0.0.1:9100",
	}
	klog.Debugf("start running replicaset controller\n")
	go replicaset.NewReplicaSetController(msgConfig, clientConfig).Run(ctx)
	return nil
}
