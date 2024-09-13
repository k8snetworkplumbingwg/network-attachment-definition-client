// Copyright (c) 2021 Kubernetes Network Plumbing Working Group
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package utils

import (
	"context"
	"net"

	cnitypes "github.com/containernetworking/cni/pkg/types"
	cni100 "github.com/containernetworking/cni/pkg/types/100"

	v1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// EnsureCIDR parses/verify CIDR ip string and convert to net.IPNet
func EnsureCIDR(cidr string) *net.IPNet {
	ip, net, err := net.ParseCIDR(cidr)
	Expect(err).NotTo(HaveOccurred())
	net.IP = ip
	return net
}

var _ = Describe("Netwok Attachment Definition manipulations", func() {

	It("test convertDNS", func() {
		cniDNS := cnitypes.DNS{
			Nameservers: []string{"aaa", "bbb"},
			Domain:      "testDomain",
			Search:      []string{"1.example.com", "2.example.com"},
			Options:     []string{"debug", "inet6"},
		}

		v1DNS := convertDNS(cniDNS)
		Expect(v1DNS.Nameservers).To(Equal(cniDNS.Nameservers))
		Expect(v1DNS.Domain).To(Equal(cniDNS.Domain))
		Expect(v1DNS.Search).To(Equal(cniDNS.Search))
		Expect(v1DNS.Options).To(Equal(cniDNS.Options))
	})

	It("set network status into pod", func() {
		fakePod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "fakePod1",
				Namespace: "fakeNamespace1",
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "fakeContainer",
						Image: "fakeImage",
					},
				},
			},
		}
		fakeStatus := []v1.NetworkStatus{
			{
				Name:      "cbr0",
				Interface: "eth0",
				IPs:       []string{"10.244.1.2"},
				Mac:       "92:79:27:01:7c:ce",
				Mtu:       1500,
			},
			{
				Name:      "test-net-attach-def-1",
				Interface: "net1",
				IPs:       []string{"1.1.1.1"},
				Mac:       "ea:0e:fa:63:95:f9",
			},
		}

		clientSet := fake.NewSimpleClientset(fakePod)
		pod, err := clientSet.CoreV1().Pods("fakeNamespace1").Get(context.TODO(), "fakePod1", metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())

		err = SetNetworkStatus(clientSet, pod, fakeStatus)
		Expect(err).NotTo(HaveOccurred())

		pod, err = clientSet.CoreV1().Pods("fakeNamespace1").Get(context.TODO(), "fakePod1", metav1.GetOptions{})
		Expect(err).NotTo(HaveOccurred())
		getStatuses, err := GetNetworkStatus(pod)
		Expect(err).NotTo(HaveOccurred())
		Expect(fakeStatus).To(Equal(getStatuses))
	})

	Context("create network status from cni result", func() {
		var cniResult *cni100.Result
		var networkStatus *v1.NetworkStatus

		BeforeEach(func() {
			cniResult = &cni100.Result{
				CNIVersion: "1.0.0",
				Interfaces: []*cni100.Interface{
					{
						Name:    "net1",
						Mac:     "92:79:27:01:7c:cf",
						Sandbox: "/proc/1123/ns/net",
						Mtu:     9000,
					},
				},
				IPs: []*cni100.IPConfig{
					{
						Address: *EnsureCIDR("1.1.1.3/24"),
					},
					{
						Address: *EnsureCIDR("2001::1/64"),
					},
				},
			}
			var err error
			networkStatus, err = CreateNetworkStatus(cniResult, "test-net-attach-def", false, nil)
			Expect(err).NotTo(HaveOccurred())
		})

		It("create network status from cni result", func() {
			Expect(networkStatus.Name).To(Equal("test-net-attach-def"))
			Expect(networkStatus.Interface).To(Equal("net1"))
			Expect(networkStatus.Mtu).To(Equal(9000))
			Expect(networkStatus.Mac).To(Equal("92:79:27:01:7c:cf"))
			Expect(networkStatus.IPs).To(Equal([]string{"1.1.1.3", "2001::1"}))
		})

		It("the network status do **not** report a gateway", func() {
			Expect(networkStatus.Gateway).To(BeEmpty())
		})

		When("DeviceInfo is used as an attribute", func() {
			var deviceInfo *v1.DeviceInfo

			BeforeEach(func() {
				deviceInfo = &v1.DeviceInfo{
					Type:    "pci",
					Version: "v1.1.0",
					Pci: &v1.PciDevice{
						PciAddress:        "0000:01:02.2",
						PfPciAddress:      "0000:01:02.0",
						RepresentorDevice: "eth3",
					},
				}
				var err error
				networkStatus, err = CreateNetworkStatus(cniResult, "test-net-attach-def", false, deviceInfo)
				Expect(err).NotTo(HaveOccurred())
			})

			It("create network status from cni result", func() {
				Expect(networkStatus.DeviceInfo.Type).To(Equal("pci"))
				Expect(networkStatus.DeviceInfo.Version).To(Equal("v1.1.0"))
				Expect(networkStatus.DeviceInfo.Pci.PciAddress).To(Equal("0000:01:02.2"))
				Expect(networkStatus.DeviceInfo.Pci.PfPciAddress).To(Equal("0000:01:02.0"))
				Expect(networkStatus.DeviceInfo.Pci.RepresentorDevice).To(Equal("eth3"))
			})
		})

		When("The CNI results features routes with default route", func() {
			const gatewayIP = "10.10.10.10"
			BeforeEach(func() {
				cniResult.Routes = []*cnitypes.Route{
					{
						Dst: net.IPNet{
							IP:   net.IP{0, 0, 0, 0},
							Mask: net.CIDRMask(0, 0),
						},
						GW: net.ParseIP(gatewayIP),
					},
				}
				var err error
				networkStatus, err = CreateNetworkStatus(cniResult, "test-net-attach-def", false, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("the network status report a gateway", func() {
				Expect(networkStatus.Gateway).To(ConsistOf(gatewayIP))
			})

			It("the network status handles multiple default routes", func() {
				const secondDefaultRoute = "20.20.20.20"

				cniResult.Routes = append(cniResult.Routes, &cnitypes.Route{
					GW: net.ParseIP(secondDefaultRoute),
				})
				networkStatus, err := CreateNetworkStatus(cniResult, "test-net-attach-def", false, nil)
				Expect(err).NotTo(HaveOccurred())
				Expect(networkStatus.Gateway).To(ConsistOf(gatewayIP, secondDefaultRoute))
			})
		})

		When("The CNI results features routes that are **not** the default route", func() {
			BeforeEach(func() {
				cniResult.Routes = []*cnitypes.Route{
					{
						Dst: net.IPNet{
							IP:   net.IP{10, 10, 10, 0},
							Mask: net.CIDRMask(24, 32),
						},
						GW: net.IP{10, 10, 10, 10},
					},
				}
				var err error
				networkStatus, err = CreateNetworkStatus(cniResult, "test-net-attach-def", false, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("the network status **should not** report a gateway", func() {
				Expect(networkStatus.Gateway).To(BeEmpty())
			})
		})
	})

	Context("create network statuses from CNI result with multiple interfaces", func() {
		var cniResult *cni100.Result
		var networkStatuses []*v1.NetworkStatus

		BeforeEach(func() {
			cniResult = &cni100.Result{
				CNIVersion: "1.1.0",
				Interfaces: []*cni100.Interface{
					{
						Name: "foo",
						Mac:  "00:AA:BB:CC:DD:33",
					},
					{
						Name:    "example0",
						Mac:     "00:AA:BB:CC:DD:01",
						Sandbox: "/path/to/network/namespace",
						Mtu:     1500,
					},
					{
						Name:    "example1",
						Mac:     "00:AA:BB:CC:DD:02",
						Sandbox: "/path/to/network/namespace",
						Mtu:     1500,
					},
				},
				IPs: []*cni100.IPConfig{
					{
						Address:   *EnsureCIDR("192.0.2.1/24"),
						Interface: &[]int{1}[0],
					},
					{
						Address:   *EnsureCIDR("192.0.2.2/24"),
						Interface: &[]int{1}[0],
					},
					{
						Address:   *EnsureCIDR("192.0.2.3/24"),
						Interface: &[]int{2}[0],
					},
				},
				DNS: cnitypes.DNS{
					Nameservers: []string{"8.8.8.8", "8.8.4.4"},
					Search:      []string{"example.com"},
					Options:     []string{"ndots:2"},
				},
			}

		})

		Context("for a secondary network", func() {
			BeforeEach(func() {
				var err error
				networkStatuses, err = CreateNetworkStatuses(cniResult, "test-multi-net-attach-def", false, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("creates network statuses for interfaces with sandbox", func() {
				Expect(networkStatuses).To(HaveLen(2))
				// Check details for the first returned network status
				Expect(networkStatuses[0].Name).To(Equal("test-multi-net-attach-def"))
				Expect(networkStatuses[0].Interface).To(Equal("example0"))
				Expect(networkStatuses[0].Mtu).To(Equal(1500))
				Expect(networkStatuses[0].Mac).To(Equal("00:AA:BB:CC:DD:01"))
				Expect(networkStatuses[0].IPs).To(ConsistOf("192.0.2.1", "192.0.2.2"))
				Expect(networkStatuses[0].Default).To(BeFalse())

				// Check details for the second returned network status
				Expect(networkStatuses[1].Interface).To(Equal("example1"))
				Expect(networkStatuses[1].Mtu).To(Equal(1500))
				Expect(networkStatuses[1].Mac).To(Equal("00:AA:BB:CC:DD:02"))
				Expect(networkStatuses[1].IPs).To(ConsistOf("192.0.2.3"))
				Expect(networkStatuses[1].Default).To(BeFalse())
			})

			It("excludes interfaces without a sandbox", func() {
				for _, status := range networkStatuses {
					Expect(status.Interface).NotTo(Equal("foo"))
				}
			})
		})

		Context("for the cluster default network", func() {
			BeforeEach(func() {
				var err error
				networkStatuses, err = CreateNetworkStatuses(cniResult, "test-multi-net-attach-def", true, nil)
				Expect(err).NotTo(HaveOccurred())
			})

			It("creates network statuses for interfaces with sandbox", func() {
				Expect(networkStatuses).To(HaveLen(2))
				// Check details for the first returned network status
				Expect(networkStatuses[0].Name).To(Equal("test-multi-net-attach-def"))
				Expect(networkStatuses[0].Interface).To(Equal("example0"))
				Expect(networkStatuses[0].Mtu).To(Equal(1500))
				Expect(networkStatuses[0].Mac).To(Equal("00:AA:BB:CC:DD:01"))
				Expect(networkStatuses[0].IPs).To(ConsistOf("192.0.2.1", "192.0.2.2"))
				Expect(networkStatuses[0].Default).To(BeTrue())

				// Check details for the second returned network status
				Expect(networkStatuses[1].Interface).To(Equal("example1"))
				Expect(networkStatuses[1].Mtu).To(Equal(1500))
				Expect(networkStatuses[1].Mac).To(Equal("00:AA:BB:CC:DD:02"))
				Expect(networkStatuses[1].IPs).To(ConsistOf("192.0.2.3"))
				Expect(networkStatuses[1].Default).To(BeFalse())
			})

			It("excludes interfaces without a sandbox", func() {
				for _, status := range networkStatuses {
					Expect(status.Interface).NotTo(Equal("foo"))
				}
			})
		})
	})

	It("parse network selection element in pod", func() {
		selectionElement := `
		[{
			"name": "test-net-attach-def",
			"interface": "test1"
		}]`
		expectedElement := []*v1.NetworkSelectionElement{
			{
				Name:             "test-net-attach-def",
				InterfaceRequest: "test1",
				Namespace:        "fakeNamespace1",
			},
		}

		fakePod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "fakePod1",
				Namespace: "fakeNamespace1",
				Annotations: map[string]string{
					"k8s.v1.cni.cncf.io/networks": selectionElement,
				},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "fakeContainer",
						Image: "fakeImage",
					},
				},
			},
		}
		elem, err := ParsePodNetworkAnnotation(fakePod)
		Expect(err).NotTo(HaveOccurred())
		Expect(elem).To(Equal(expectedElement))
	})
})
