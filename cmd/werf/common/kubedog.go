package common

import (
	"context"
	"flag"
	"io/ioutil"

	"k8s.io/klog"
	klog_v2 "k8s.io/klog/v2"

	"github.com/werf/kubedog/pkg/display"
	"github.com/werf/logboek"
)

func InitKubedog(ctx context.Context) error {
	// set flag.Parsed() for glog
	flag.CommandLine.Parse([]string{})

	display.SetOut(logboek.Context(ctx).OutStream())
	display.SetErr(logboek.Context(ctx).ErrStream())

	if err := SilenceKlog(ctx); err != nil {
		return err
	}
	if err := SilenceKlogV2(ctx); err != nil {
		return err
	}

	return nil
}

func SilenceKlogV2(ctx context.Context) error {
	fs := flag.NewFlagSet("klog", flag.PanicOnError)
	klog_v2.InitFlags(fs)

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
	klog_v2.SetOutputBySeverity("INFO", ioutil.Discard)
	klog_v2.SetOutputBySeverity("WARNING", ioutil.Discard)
	klog_v2.SetOutputBySeverity("ERROR", ioutil.Discard)
	klog_v2.SetOutputBySeverity("FATAL", logboek.Context(ctx).ErrStream())

	return nil
}

func SilenceKlog(ctx context.Context) error {
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
	klog.SetOutputBySeverity("FATAL", logboek.Context(ctx).ErrStream())

	return nil
}
