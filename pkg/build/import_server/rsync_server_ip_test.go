package import_server

import (
	"net/netip"

	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	ginkgo "github.com/onsi/ginkgo/v2"
	gomega "github.com/onsi/gomega"
)

var _ = ginkgo.DescribeTable("bridgeNetworkIPAddress", func(settings *dockercontainer.NetworkSettings, expectedAddress, expectedError string) {
	address, err := bridgeNetworkIPAddress(settings)
	if expectedError != "" {
		gomega.Expect(err).To(gomega.MatchError(gomega.ContainSubstring(expectedError)))
		gomega.Expect(address).To(gomega.BeEmpty())
		return
	}
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	gomega.Expect(address).To(gomega.Equal(expectedAddress))
},
	ginkgo.Entry("no network settings", nil, "", "no network settings available in inspect"),
	ginkgo.Entry("no bridge network",
		&dockercontainer.NetworkSettings{Networks: map[string]*network.EndpointSettings{
			"host": {IPAddress: netip.MustParseAddr("172.17.0.2")},
		}},
		"", "not attached to the bridge network"),
	ginkgo.Entry("nil bridge endpoint",
		&dockercontainer.NetworkSettings{Networks: map[string]*network.EndpointSettings{"bridge": nil}},
		"", "not attached to the bridge network"),
	ginkgo.Entry("bridge without an ip address",
		&dockercontainer.NetworkSettings{Networks: map[string]*network.EndpointSettings{"bridge": {}}},
		"", "reports no ip address"),
	ginkgo.Entry("bridge with an ip address",
		&dockercontainer.NetworkSettings{Networks: map[string]*network.EndpointSettings{
			"bridge": {IPAddress: netip.MustParseAddr("172.17.0.2")},
		}},
		"172.17.0.2", ""),
)
