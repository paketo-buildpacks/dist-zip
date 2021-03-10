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

package distzip_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/dist-zip/distzip"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx    libcnb.DetectContext
		detect distzip.Detect
	)

	it.Before(func() {
		var err error

		ctx.Application.Path, err = ioutil.TempDir("", "dist-zip")
		Expect(err).NotTo(HaveOccurred())

		ctx.Buildpack.Metadata = map[string]interface{}{
			"configurations": []map[string]interface{}{
				{
					"name":    "BP_APPLICATION_SCRIPT",
					"default": "*/bin/*",
				},
			},
		}
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Application.Path)).To(Succeed())
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
			Expect(ioutil.WriteFile(filepath.Join(ctx.Application.Path, "app", "bin", "script-1"), []byte{}, 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(ctx.Application.Path, "app", "bin", "script-2"), []byte{}, 0755)).To(Succeed())
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
							{Name: "jre", Metadata: map[string]interface{}{"launch": true}},
							{Name: "jvm-application-package"},
							{Name: "jvm-application"},
						},
					},
				},
			}))
		})
	})

	context("single application script", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(ctx.Application.Path, "app", "bin"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(ctx.Application.Path, "app", "bin", "script"), []byte{}, 0755)).To(Succeed())
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
							{Name: "jre", Metadata: map[string]interface{}{"launch": true}},
							{Name: "jvm-application-package"},
							{Name: "jvm-application"},
						},
					},
				},
			}))
		})
	})
}
