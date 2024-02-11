/*
 * Copyright 2018-2024 the original author or authors.
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

package distzip_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/dist-zip/v5/distzip"
	"github.com/paketo-buildpacks/libpak/bard"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		buf    *bytes.Buffer
		ctx    libcnb.DetectContext
		detect distzip.Detect
	)

	it.Before(func() {
		var err error

		ctx.Application.Path = t.TempDir()
		Expect(err).NotTo(HaveOccurred())

		ctx.Buildpack.Metadata = map[string]interface{}{
			"configurations": []map[string]interface{}{
				{
					"name":    "BP_APPLICATION_SCRIPT",
					"default": "*/bin/*",
				},
			},
		}

		buf = &bytes.Buffer{}
		detect = distzip.Detect{
			Logger: bard.NewLoggerWithOptions(io.Discard, bard.WithDebug(buf))}
	})

	context("application script not found", func() {
		it("requires jvm-application-package", func() {
			Expect(detect.Detect(ctx)).To(Equal(libcnb.DetectResult{
				Pass: true,
				Plans: []libcnb.BuildPlan{
					{
						Provides: []libcnb.BuildPlanProvide{
							{Name: "jvm-application"},
						},
						Requires: []libcnb.BuildPlanRequire{
							{Name: "syft"},
							{Name: "jre", Metadata: map[string]interface{}{"launch": true}},
							{Name: "jvm-application-package"},
							{Name: "jvm-application"},
						},
					},
				},
			}))
		})
	})

	context("multiple application scripts", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(ctx.Application.Path, "app", "bin"), 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(ctx.Application.Path, "app", "bin", "script-1"), []byte{}, 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(ctx.Application.Path, "app", "bin", "script-2"), []byte{}, 0755)).To(Succeed())
		})

		it("requires jvm-application-package", func() {
			Expect(detect.Detect(ctx)).To(Equal(libcnb.DetectResult{
				Pass: true,
				Plans: []libcnb.BuildPlan{
					{
						Provides: []libcnb.BuildPlanProvide{
							{Name: "jvm-application"},
						},
						Requires: []libcnb.BuildPlanRequire{
							{Name: "syft"},
							{Name: "jre", Metadata: map[string]interface{}{"launch": true}},
							{Name: "jvm-application-package"},
							{Name: "jvm-application"},
						},
					},
				},
			}))
			Expect(buf.String()).To(ContainSubstring(fmt.Sprintf(`too many application scripts in */bin/*, candidates: [%s %s]`,
				filepath.Join(ctx.Application.Path, "app", "bin", "script-1"),
				filepath.Join(ctx.Application.Path, "app", "bin", "script-2"))))
			Expect(buf.String()).To(ContainSubstring("set a more strict `$BP_APPLICATION_SCRIPT` pattern that only matches a single script"))
		})
	})

	context("single application script", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(ctx.Application.Path, "app", "bin"), 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(ctx.Application.Path, "app", "bin", "script"), []byte{}, 0755)).To(Succeed())
		})

		it("requires and provides jvm-application-package", func() {
			Expect(detect.Detect(ctx)).To(Equal(libcnb.DetectResult{
				Pass: true,
				Plans: []libcnb.BuildPlan{
					{
						Provides: []libcnb.BuildPlanProvide{
							{Name: "jvm-application"},
							{Name: "jvm-application-package"},
						},
						Requires: []libcnb.BuildPlanRequire{
							{Name: "syft"},
							{Name: "jre", Metadata: map[string]interface{}{"launch": true}},
							{Name: "jvm-application-package"},
							{Name: "jvm-application"},
						},
					},
				},
			}))
		})
	})

	context("$BP_LIVE_RELOAD_ENABLED is set", func() {
		it.Before(func() {
			t.Setenv("BP_LIVE_RELOAD_ENABLED", "true")

			Expect(os.MkdirAll(filepath.Join(ctx.Application.Path, "app", "bin"), 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(ctx.Application.Path, "app", "bin", "script"), []byte{}, 0755)).To(Succeed())
		})

		it("requires watchexec", func() {
			Expect(detect.Detect(ctx)).To(Equal(libcnb.DetectResult{
				Pass: true,
				Plans: []libcnb.BuildPlan{
					{
						Provides: []libcnb.BuildPlanProvide{
							{Name: "jvm-application"},
							{Name: "jvm-application-package"},
						},
						Requires: []libcnb.BuildPlanRequire{
							{Name: "syft"},
							{Name: "jre", Metadata: map[string]interface{}{"launch": true}},
							{Name: "jvm-application-package"},
							{Name: "jvm-application"},
							{Name: "watchexec"},
						},
					},
				},
			}))
		})
	})
}
