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
	"github.com/paketo-buildpacks/libpak/bard"
)

type Build struct {
	Logger bard.Logger
}

func (b Build) Build(context libcnb.BuildContext) (libcnb.BuildResult, error) {
	sr := ScriptResolver{ApplicationPath: context.Application.Path}
	s, ok, err := sr.Resolve()
	if err != nil {
		return libcnb.BuildResult{}, fmt.Errorf("unable to detect application scripts\n%w", err)
	}

	if !ok {
		return libcnb.BuildResult{}, nil
	}

	b.Logger.Title(context.Buildpack)
	b.Logger.Body(bard.FormatUserConfig("BP_APPLICATION_SCRIPT", "the application start script", DefaultPattern))

	result := libcnb.BuildResult{
		Processes: []libcnb.Process{
			{Type: "dist-zip", Command: s},
			{Type: "task", Command: s},
			{Type: "web", Command: s},
		},
	}

	return result, nil
}
