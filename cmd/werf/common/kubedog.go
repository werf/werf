package common

import (
	"flag"
	"io/ioutil"

	"github.com/werf/kubedog/pkg/display"
	"github.com/werf/logboek"
	"k8s.io/klog"
)

func InitKubedog() error {
	// set flag.Parsed() for glog
	flag.CommandLine.Parse([]string{})

	// Suppress info and warnings from client-go reflector
	klog.SetOutputBySeverity("INFO", ioutil.Discard)
	klog.SetOutputBySeverity("WARNING", ioutil.Discard)
	klog.SetOutputBySeverity("ERROR", logboek.GetErrStream())
	klog.SetOutputBySeverity("FATAL", logboek.GetErrStream())

	display.SetOut(logboek.GetOutStream())
	display.SetErr(logboek.GetErrStream())

	return nil
}
