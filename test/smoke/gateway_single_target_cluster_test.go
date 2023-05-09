package smoke

import (
	"context"
	"strings"
	"time"

	"github.com/davecgh/go-spew/spew"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayapi "sigs.k8s.io/gateway-api/apis/v1beta1"
)

var _ = Describe("Gateway single target cluster", func() {

	// gwname is the name of the Gateway that will be used
	// as the test subject. It is a generated name to allow for
	// test parallelism
	var gwname string
	var ctx context.Context = context.Background()

	BeforeEach(func() {

		// NOTE This will only be useful once we have multi-teanancy and can create gateways
		// in different tenant namespaces
		//
		// By("creating a tenant namespace in the control plane")
		// testNamespace = "test-ns-" + nameGenerator.Generate()
		// ns := &corev1.Namespace{
		// 	TypeMeta:   metav1.TypeMeta{APIVersion: "v1", Kind: "Namespace"},
		// 	ObjectMeta: metav1.ObjectMeta{Name: testNamespace},
		// }
		// err := k8sClient.Create(context.Background(), ns)
		// Expect(err).ToNot(HaveOccurred())
		// n := &corev1.Namespace{}
		// Eventually(func() bool {
		// 	err := k8sClient.Get(context.Background(), types.NamespacedName{Name: testNamespace}, n)
		// 	return err == nil
		// }, 60*time.Second, 5*time.Second).Should(BeTrue())

		By("creating a Gateway in the control plane")
		gwname = "t-smoke-" + nameGenerator.Generate()

		hostname := gatewayapi.Hostname(strings.Join([]string{gwname, managedZone}, "."))
		gw := &gatewayapi.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:        gwname,
				Namespace:   tenantNamespace,
				Annotations: map[string]string{clusterSelectorLabelKey: clusterSelectorLabelValue},
			},
			Spec: gatewayapi.GatewaySpec{
				GatewayClassName: gwClassName,
				Listeners: []gatewayapi.Listener{{
					Name:     "https",
					Hostname: &hostname,
					Port:     443,
					Protocol: gatewayapi.HTTPSProtocolType,
					TLS: &gatewayapi.GatewayTLSConfig{
						CertificateRefs: []gatewayapi.SecretObjectReference{{
							Name: gatewayapi.ObjectName(hostname),
						}},
					},
				}},
			},
		}

		err := cpClient.Create(context.Background(), gw)
		Expect(err).ToNot(HaveOccurred())

	})

	AfterEach(func() {

		gw := &gatewayapi.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      gwname,
				Namespace: tenantNamespace,
			},
		}
		err := cpClient.Delete(context.Background(), gw, client.PropagationPolicy(metav1.DeletePropagationForeground))
		Expect(err).ToNot(HaveOccurred())
	})

	When("the controller picks it up", func() {

		It("sets the 'Accepted' condition to true", func() {

			gw := &gatewayapi.Gateway{}
			Eventually(func() bool {
				if err := cpClient.Get(ctx, types.NamespacedName{Name: gwname, Namespace: tenantNamespace}, gw); err != nil {
					return false
				}
				return meta.IsStatusConditionTrue(gw.Status.Conditions, string(gatewayapi.GatewayConditionAccepted))
			}, 60*time.Second, 5*time.Second).Should(BeTrue())
		})
	})

	When("an HTTPRoute is attached ot the Gateway", func() {

		By("attaching an HTTPRoute to the Gateway in the dataplane")
		// TODO

		It("sets the 'Programmed' condition to true", func() {
			gw := &gatewayapi.Gateway{}
			Eventually(func() bool {
				if err := cpClient.Get(ctx, types.NamespacedName{Name: gwname, Namespace: tenantNamespace}, gw); err != nil {
					return false
				}
				spew.Dump(gw.Status)
				return meta.IsStatusConditionTrue(gw.Status.Conditions, string(gatewayapi.GatewayConditionProgrammed))
			}, 300*time.Second, 5*time.Second).Should(BeTrue())
		})

		It("makes available a hostname that resolves to the dataplate Gateway", func() {
			// TODO
		})

		It("makes available a hostname that is reachable by https", func() {
			// TODO
		})
	})

})
