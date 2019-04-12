/*
Copyright 2017 DigitalOcean

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	goflag "flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/apiserver/pkg/server/healthz"
	"k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"
	"k8s.io/klog"
	"k8s.io/kubernetes/cmd/cloud-controller-manager/app"
	"k8s.io/kubernetes/cmd/cloud-controller-manager/app/options"
	_ "k8s.io/kubernetes/pkg/client/metrics/prometheus" // for client metric registration
	_ "k8s.io/kubernetes/pkg/version/prometheus"        // for version metric registration
	"k8s.io/kubernetes/pkg/version/verflag"

	_ "github.com/digitalocean/digitalocean-cloud-controller-manager/cloud-controller-manager/do"
	"github.com/spf13/pflag"
)

func init() {
	healthz.DefaultHealthz()
}

func main() {
	s, err := options.NewCloudControllerManagerOptions()
	if err != nil {
		klog.Fatalf("failed to create config options: %s", err)
	}

	// s.AddFlags(pflag.CommandLine)

	// TODO: once we switch everything over to Cobra commands, we can go back to calling
	// utilflag.InitFlags() (by removing its pflag.Parse() call). For now, we have to set the
	// normalize func and add the go flag set by hand.
	pflag.CommandLine.SetNormalizeFunc(flag.WordSepNormalizeFunc)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	logs.InitLogs()
	defer logs.FlushLogs()

	verflag.PrintAndExitIfRequested()

	// digitalocean overrides
	s.KubeCloudShared.AllowUntaggedCloud = true

	config, err := s.Config(nil, nil)
	if err != nil {
		klog.Fatalf("failed to create component config: %s", err)
	}

	stop := make(chan struct{})
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		close(stop)
	}()

	if err := app.Run(config.Complete(), stop); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
