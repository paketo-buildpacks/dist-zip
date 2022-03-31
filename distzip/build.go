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

	"github.com/paketo-buildpacks/libpak/effect"
	"github.com/paketo-buildpacks/libpak/sbom"

	"github.com/buildpacks/libcnb"
	"github.com/paketo-buildpacks/libpak"
	"github.com/paketo-buildpacks/libpak/bard"
)

type Build struct {
	Logger      bard.Logger
	SBOMScanner sbom.SBOMScanner
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	result := libcnb.NewBuildResult()

	cr, err := libpak.NewConfigurationResolver(context.Buildpack, nil)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create configuration resolver\n%w", err)
	}

	sr := ScriptResolver{
		ApplicationPath:       context.Application.Path,
		ConfigurationResolver: cr,
		Logger:                b.Logger,
	}
	s, ok, err := sr.Resolve()
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to detect application scripts\n%w", err)
	}

	if !ok {
		for _, entry := range context.Plan.Entries {
			result.Unmet = append(result.Unmet, libcnb.UnmetPlanEntry{Name: entry.Name})
		}
		return result, nil
	}

	b.Logger.Title(context.Buildpack)

	_, err = libpak.NewConfigurationResolver(context.Buildpack, &b.Logger)
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create configuration resolver\n%w", err)
	}

	result.Processes = append(result.Processes,
		libcnb.Process{Type: "dist-zip", Command: s},
		libcnb.Process{Type: "task", Command: s},
		libcnb.Process{Type: "web", Command: s, Default: true},
	)

	if cr.ResolveBool("BP_LIVE_RELOAD_ENABLED") {
		for i := 0; i < len(result.Processes); i++ {
			result.Processes[i].Default = false
		}

		result.Processes = append(result.Processes,
			libcnb.Process{
				Type:      "reload",
				Command:   "watchexec",
				Arguments: []string{"-r", s},
				Direct:    false,
				Default:   true,
			},
		)
	}

	if b.SBOMScanner == nil {
		b.SBOMScanner = sbom.NewSyftCLISBOMScanner(context.Layers, effect.NewExecutor(), b.Logger)
	}
	if err := b.SBOMScanner.ScanLaunch(context.Application.Path, libcnb.SyftJSON, libcnb.CycloneDXJSON); err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to create Build SBoM \n%w", err)
	}

	return result, nil
}
