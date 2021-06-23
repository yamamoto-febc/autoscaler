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
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/sacloud/libsacloud/v2/sacloud"
)

type ResourceDefRouter struct {
	*ResourceDefBase `yaml:",inline"`
	Plans            []*RouterPlan `yaml:"plans"`
}

func (d *ResourceDefRouter) resourcePlans() ResourcePlans {
	if len(d.Plans) == 0 {
		return DefaultRouterPlans
	}
	var plans ResourcePlans
	for _, p := range d.Plans {
		plans = append(plans, p)
	}
	return plans
}

func (d *ResourceDefRouter) Validate(ctx context.Context, apiClient sacloud.APICaller) []error {
	errors := &multierror.Error{}

	selector := d.Selector()
	if selector == nil {
		errors = multierror.Append(errors, fmt.Errorf("selector: required"))
	} else {
		if selector.Zone == "" {
			errors = multierror.Append(errors, fmt.Errorf("selector.Zone: required"))
		} else {
			if errs := d.validatePlans(ctx, apiClient); len(errs) > 0 {
				errors = multierror.Append(errors, errs...)
			}

			resources, err := d.findCloudResources(ctx, apiClient)
			if err != nil {
				errors = multierror.Append(errors, err)
			}
			if len(d.children) > 0 && len(resources) > 1 {
				var names []string
				for _, r := range resources {
					names = append(names, fmt.Sprintf("{Zone:%s, ID:%s, Name:%s}", selector.Zone, r.ID, r.Name))
				}
				errors = multierror.Append(errors,
					fmt.Errorf("A resource definition with children must return one resource, but got multiple resources: definition: {Type:%s, Selector:%s}, got: %s",
						d.Type(), d.Selector(), fmt.Sprintf("[%s]", strings.Join(names, ",")),
					))
			}
		}
	}

	// set prefix
	errors = multierror.Prefix(errors, fmt.Sprintf("resource=%s:", d.Type().String())).(*multierror.Error)
	return errors.Errors
}

func (d *ResourceDefRouter) validatePlans(ctx context.Context, apiClient sacloud.APICaller) []error {
	if len(d.Plans) > 0 {
		if len(d.Plans) == 1 {
			return []error{fmt.Errorf("at least two plans must be specified")}
		}

		availablePlans, err := sacloud.NewInternetPlanOp(apiClient).Find(ctx, d.Selector().Zone, nil)
		if err != nil {
			return []error{fmt.Errorf("validating router plan failed: %s", err)}
		}

		// for unique check: plan name
		names := map[string]struct{}{}

		errors := &multierror.Error{}
		for _, p := range d.Plans {
			if p.Name != "" {
				if _, ok := names[p.Name]; ok {
					errors = multierror.Append(errors, fmt.Errorf("plan name %q is duplicated", p.Name))
				}
				names[p.Name] = struct{}{}
			}

			exists := false
			for _, available := range availablePlans.InternetPlans {
				if available.Availability.IsAvailable() && available.BandWidthMbps == p.BandWidth {
					exists = true
					break
				}
			}
			if !exists {
				errors = multierror.Append(errors, fmt.Errorf("plan{band_width:%d} not exists", p.BandWidth))
			}
		}
		return errors.Errors
	}
	return nil
}

func (d *ResourceDefRouter) Compute(ctx *RequestContext, apiClient sacloud.APICaller) (Resources, error) {
	cloudResources, err := d.findCloudResources(ctx, apiClient)
	if err != nil {
		return nil, err
	}

	var resources Resources
	for _, router := range cloudResources {
		r, err := NewResourceRouter(ctx, apiClient, d, d.Selector().Zone, router)
		if err != nil {
			return nil, err
		}
		resources = append(resources, r)
	}
	return resources, nil
}

func (d *ResourceDefRouter) findCloudResources(ctx context.Context, apiClient sacloud.APICaller) ([]*sacloud.Internet, error) {
	routerOp := sacloud.NewInternetOp(apiClient)
	selector := d.Selector()

	found, err := routerOp.Find(ctx, selector.Zone, selector.findCondition())
	if err != nil {
		return nil, fmt.Errorf("computing state failed: %s", err)
	}
	if len(found.Internet) == 0 {
		return nil, fmt.Errorf("resource not found with selector: %s", selector.String())
	}

	return found.Internet, nil
}
