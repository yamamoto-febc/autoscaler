// Copyright 2021 The sacloud Authors
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

package core

import (
	"fmt"

	"github.com/goccy/go-yaml"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type ResourceGroup struct {
	HandlerConfigs []*ResourceHandlerConfig `yaml:"handlers"`
	Resources      Resources                `yaml:"resources"`
	Name           string
}

type ResourceHandlerConfig struct {
	Name string `yaml:"name"`
	// TODO 未実装
	//Selector *ResourceSelector `yaml:"selector"`
}

func (rg *ResourceGroup) UnmarshalYAML(data []byte) error {
	var rawMap map[string]interface{}
	if err := yaml.Unmarshal(data, &rawMap); err != nil {
		return err
	}

	resourceGroup := &ResourceGroup{}
	resources := rawMap["resources"].([]interface{})
	for _, rawResource := range resources {
		v := rawResource.(map[string]interface{})
		resource, err := rg.unmarshalResourceFromMap(v)
		if err != nil {
			return err
		}

		rg.setParentResource(nil, resource)
		resourceGroup.Resources = append(resourceGroup.Resources, resource)
	}

	if rawHandlers, ok := rawMap["handlers"]; ok {
		handlers := rawHandlers.([]interface{})
		for _, name := range handlers {
			if n, ok := name.(string); ok {
				resourceGroup.HandlerConfigs = append(resourceGroup.HandlerConfigs, &ResourceHandlerConfig{Name: n})
			}
		}
	}

	*rg = *resourceGroup
	return nil
}

func (rg *ResourceGroup) unmarshalResourceFromMap(data map[string]interface{}) (Resource, error) {
	rawTypeName, ok := data["type"]
	if !ok {
		return nil, fmt.Errorf("'type' field required: %v", data)
	}
	typeName, ok := rawTypeName.(string)
	if !ok {
		return nil, fmt.Errorf("'type' is not string: %v", data)
	}

	remarshelded, err := yaml.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("yaml.Marshal failed with %v", data)
	}

	var resources Resources
	if rawChildren, ok := data["resources"]; ok {
		if children, ok := rawChildren.([]interface{}); ok {
			for _, child := range children {
				if c, ok := child.(map[string]interface{}); ok {
					r, err := rg.unmarshalResourceFromMap(c)
					if err != nil {
						return nil, err
					}
					resources = append(resources, r)
				}
			}
		}
	}

	var resource Resource
	switch typeName {
	case "Server":
		v := &Server{}
		if err := yaml.Unmarshal(remarshelded, v); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal failed with %v", data)
		}
		v.Children = resources
		resource = v
	case "ServerGroup":
		v := &ServerGroup{}
		if err := yaml.Unmarshal(remarshelded, v); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal failed with %v", data)
		}
		v.Children = resources
		resource = v
	case "EnhancedLoadBalancer", "ELB":
		v := &EnhancedLoadBalancer{}
		if err := yaml.Unmarshal(remarshelded, v); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal failed with %v", data)
		}
		// TypeNameのエイリアスを正規化
		v.TypeName = "EnhancedLoadBalancer"
		v.Children = resources
		resource = v
	case "GSLB":
		v := &GSLB{}
		if err := yaml.Unmarshal(remarshelded, v); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal failed with %v", data)
		}
		v.Children = resources
		resource = v
	case "DNS":
		v := &DNS{}
		if err := yaml.Unmarshal(remarshelded, v); err != nil {
			return nil, fmt.Errorf("yaml.Unmarshal failed with %v", data)
		}
		v.Children = resources
		resource = v
	default:
		return nil, fmt.Errorf("received unexpected type: %s", typeName)
	}

	return resource, nil
}

func (rg *ResourceGroup) setParentResource(parent, r Resource) {
	if parent != nil {
		if v, ok := r.(ChildResource); ok {
			v.SetParent(parent)
		}
	}
	for _, child := range r.Resources() {
		rg.setParentResource(r, child)
	}
}

func (rg *ResourceGroup) HandleAll(ctx *Context, apiClient sacloud.APICaller, handlerFilters Handlers) error {
	handlers, err := rg.handlers(handlerFilters)
	if err != nil {
		return err
	}
	return rg.handleAll(ctx, apiClient, handlers)
}

func (rg *ResourceGroup) handleAll(ctx *Context, apiClient sacloud.APICaller, handlers Handlers) error {
	allComputed, err := rg.computeAll(ctx, apiClient)
	if err != nil {
		return err
	}

	// preHandle
	if err := rg.handleAllByFunc(allComputed, handlers, func(h *Handler, c Computed) error {
		return h.PreHandle(ctx, c)
	}); err != nil {
		return err
	}

	// handle
	if err := rg.handleAllByFunc(allComputed, handlers, func(h *Handler, c Computed) error {
		return h.Handle(ctx, c)
	}); err != nil {
		return err
	}

	// refresh
	allComputed, err = rg.computeAll(ctx.ForRefresh(), apiClient) // refresh
	if err != nil {
		return err
	}

	// postHandle
	if err := rg.handleAllByFunc(allComputed, handlers, func(h *Handler, c Computed) error {
		return h.PostHandle(ctx, c)
	}); err != nil {
		return err
	}

	rg.clearCacheAll()
	return nil
}

func (rg *ResourceGroup) handleAllByFunc(allComputed []Computed, handlers Handlers, fn func(*Handler, Computed) error) error {
	for _, computed := range allComputed {
		for _, handler := range handlers {
			if err := fn(handler, computed); err != nil {
				return err
			}
		}
	}
	return nil
}

func (rg *ResourceGroup) computeAll(ctx *Context, apiClient sacloud.APICaller) ([]Computed, error) {
	// TODO 並列化
	var allComputed []Computed
	err := rg.Resources.Walk(func(resource Resource) error {
		computed, err := resource.Compute(ctx, apiClient)
		if err != nil {
			return err
		}
		allComputed = append(allComputed, computed...)
		return nil
	})
	return allComputed, err
}

func (rg *ResourceGroup) clearCacheAll() {
	rg.Resources.Walk(func(resource Resource) error { // nolint 戻り値のerrorを無視しているがerrorが返ることはない
		resource.ClearCache()
		return nil
	})
}

// Handlers 引数で指定されたハンドラーのリストをHandlerConfigsに合致するハンドラだけにフィルタして返す
func (rg *ResourceGroup) handlers(allHandlers Handlers) (Handlers, error) {
	if len(rg.HandlerConfigs) == 0 {
		return allHandlers, nil
	}
	var handlers Handlers
	for _, conf := range rg.HandlerConfigs {
		var found *Handler
		for _, h := range allHandlers {
			if h.Name == conf.Name {
				found = h
				break
			}
		}
		if found == nil {
			return nil, fmt.Errorf("handler %q not found", conf.Name)
		}
		handlers = append(handlers, found)
	}
	return handlers, nil
}
