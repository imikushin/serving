/*
 Copyright 2019 The Knative Authors

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

package resources

import (
	"github.com/knative/pkg/apis/istio/v1alpha3"
	"github.com/knative/pkg/kmeta"
	"github.com/knative/serving/pkg/apis/serving"
	v1alpha12 "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	"github.com/knative/serving/pkg/reconciler/v1alpha1/revision/resources/names"
	"github.com/knative/serving/pkg/system"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MakeVirtualService creates an Istio VirtualService for the Revision.
func MakeVirtualService(rev *v1alpha12.Revision) *v1alpha3.VirtualService {
	vs := &v1alpha3.VirtualService{
		ObjectMeta: metav1.ObjectMeta{
			Name:            rev.Name,
			OwnerReferences: []metav1.OwnerReference{*kmeta.NewControllerRef(rev)},
			Annotations:     rev.ObjectMeta.Annotations,
		},
		Spec: *makeVirtualServiceSpec(rev),
	}

	// Populate the Revision labels.
	if vs.Labels == nil {
		vs.Labels = make(map[string]string)
	}
	vs.Labels[serving.RevisionLabelKey] = rev.Labels[serving.RevisionLabelKey]
	return vs
}

func makeVirtualServiceSpec(rev *v1alpha12.Revision) *v1alpha3.VirtualServiceSpec {
	return &v1alpha3.VirtualServiceSpec{
		// TODO(imikushin) get knative gateways from the context, ideally just use "mesh"
		Gateways: []string{
			"knative-ingress-gateway.knative-serving.svc.cluster.local",
			"knative-shared-gateway.knative-serving.svc.cluster.local",
			"mesh"},
		Hosts:    getHosts(rev),
		Http:     []v1alpha3.HTTPRoute{*makeHTTPRoute(rev)},
	}
}

func makeHTTPRoute(rev *v1alpha12.Revision) *v1alpha3.HTTPRoute {
	return &v1alpha3.HTTPRoute{
		Route: []v1alpha3.HTTPRouteDestination{{
			Destination: v1alpha3.Destination{
				Host: names.ActivatorOrInternalK8sServiceHost(rev),
				Port: v1alpha3.PortSelector{Number: uint32(80)},
			},
			Weight: 100,
		}},
		Retries: &v1alpha3.HTTPRetry{
			Attempts:      1000,
			PerTryTimeout: "100ms",
			// RetryOn: "gateway-error", // TODO(imikushin) enable this: currently not supported by our installed istio
		},
		Mirror: &v1alpha3.Destination{
			Host: "activator-service." + system.Namespace + ".svc.cluster.local",
			Port: v1alpha3.PortSelector{Number: uint32(80)},
		},
	}
}

func getHosts(rev *v1alpha12.Revision) []string {
	return []string{names.K8sServiceHost(rev)}
}
