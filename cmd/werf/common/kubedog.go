package common

import (
	"context"
	"flag"
	"io/ioutil"

	"k8s.io/klog"

	"github.com/werf/kubedog/pkg/display"
	"github.com/werf/logboek"
)

func InitKubedog(ctx context.Context) error {
	// set flag.Parsed() for glog
	flag.CommandLine.Parse([]string{})

	// Suppress info and warnings from client-go reflector
	klog.SetOutputBySeverity("INFO", ioutil.Discard)
	klog.SetOutputBySeverity("WARNING", ioutil.Discard)
	klog.SetOutputBySeverity("ERROR", logboek.Context(ctx).ProxyErrStream())
	klog.SetOutputBySeverity("FATAL", logboek.Context(ctx).ProxyErrStream())

	display.SetOut(logboek.Context(ctx).ProxyOutStream())
	display.SetErr(logboek.Context(ctx).ProxyErrStream())

	return nil
}
