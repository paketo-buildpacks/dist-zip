/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package distzip

import (
	"fmt"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
)

const (
	PlanEntryJVMApplication        = "jvm-application"
	PlanEntryJVMApplicationPackage = "jvm-application-package"
	PlanEntryJRE                   = "jre"
	PlanEntryWatchexec             = "watchexec"
	PlanEntrySyft                  = "syft"
)

type Detect struct{}

func (Detect) Detect(context libcnb.DetectContext) (libcnb.DetectResult, error) {
	result := libcnb.DetectResult{
		Pass: true,
		Plans: []libcnb.BuildPlan{
			{
				Provides: []libcnb.BuildPlanProvide{
					{Name: PlanEntryJVMApplication},
				},
				Requires: []libcnb.BuildPlanRequire{
					{Name: PlanEntrySyft},
					{Name: PlanEntryJRE, Metadata: map[string]interface{}{"launch": true}},
					{Name: PlanEntryJVMApplicationPackage},
					{Name: PlanEntryJVMApplication},
				},
			},
		},
	}

	cr, err := libpak.NewConfigurationResolver(context.Buildpack, nil)
	if err != nil {
		return libcnb.DetectResult{}, fmt.Errorf("unable to create configuration resolver\n%w", err)
	}

	sr := ScriptResolver{
		ApplicationPath:       context.Application.Path,
		ConfigurationResolver: cr,
	}
	if _, ok, _ := sr.Resolve(); ok {
		result.Plans[0].Provides = append(result.Plans[0].Provides, libcnb.BuildPlanProvide{Name: PlanEntryJVMApplicationPackage})
	}

	if cr.ResolveBool("BP_LIVE_RELOAD_ENABLED") {
		for i := range result.Plans {
			result.Plans[i].Requires = append(result.Plans[i].Requires, libcnb.BuildPlanRequire{
				Name: PlanEntryWatchexec,
			})
		}
	}

	return result, nil
}
