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

	fs := flag.NewFlagSet("klog", flag.PanicOnError)
	klog.InitFlags(fs)
	if err := fs.Set("logtostderr", "false"); err != nil {
		return err
	}
	if err := fs.Set("alsologtostderr", "false"); err != nil {
		return err
	}
	if err := fs.Set("stderrthreshold", "5"); err != nil {
		return err
	}

	// Suppress info and warnings from client-go reflector
	klog.SetOutputBySeverity("INFO", ioutil.Discard)
	klog.SetOutputBySeverity("WARNING", ioutil.Discard)
	klog.SetOutputBySeverity("ERROR", ioutil.Discard)
	klog.SetOutputBySeverity("FATAL", logboek.Context(ctx).ProxyErrStream())

	display.SetOut(logboek.Context(ctx).ProxyOutStream())
	display.SetErr(logboek.Context(ctx).ProxyErrStream())

	return nil
}
