// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package istio

import (
	"context"
	"path/filepath"

	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/gardener/gardener/pkg/operation/botanist/component"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type proxyProtocol struct {
	namespace    string
	chartApplier kubernetes.ChartApplier
	client       client.Client
	chartPath    string
}

// NewProxyProtocolGateway creates a new DeployWaiter for istio which
// adds a PROXY Protocol listener to the istio-ingressgateway.
func NewProxyProtocolGateway(
	namespace string,
	chartApplier kubernetes.ChartApplier,
	client client.Client,
	chartsRootPath string,
) component.DeployWaiter {
	return &proxyProtocol{
		namespace:    namespace,
		chartApplier: chartApplier,
		client:       client,
		chartPath:    filepath.Join(chartsRootPath, istioReleaseName, "istio-proxy-protocol"),
	}
}

func (i *proxyProtocol) Deploy(ctx context.Context) error {
	return i.chartApplier.Apply(ctx, i.chartPath, i.namespace, istioReleaseName)
}

func (i *proxyProtocol) Destroy(ctx context.Context) error {
	objs := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"apiVersion": "networking.istio.io/v1alpha3",
				"kind":       "EnvoyFilter",
				"metadata": map[string]interface{}{
					"name":      "proxy-protocol",
					"namespace": i.namespace,
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "networking.istio.io/v1beta1",
				"kind":       "Gateway",
				"metadata": map[string]interface{}{
					"name":      "proxy-protocol",
					"namespace": i.namespace,
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "networking.istio.io/v1alpha3",
				"kind":       "EnvoyFilter",
				"metadata": map[string]interface{}{
					"name":      "proxy-protocol-blackhole",
					"namespace": i.namespace,
				},
			},
		},
	}

	for _, obj := range objs {
		if err := i.client.Delete(ctx, obj); err != nil {
			if !apierrors.IsNotFound(err) && !meta.IsNoMatchError(err) {
				return err
			}
		}
	}

	return nil
}

func (i *proxyProtocol) Wait(ctx context.Context) error {
	return nil
}

func (i *proxyProtocol) WaitCleanup(ctx context.Context) error {
	return nil
}
