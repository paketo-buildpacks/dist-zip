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

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		ctx libcnb.BuildContext
	)

	it.Before(func() {
		var err error

		ctx.Application.Path, err = ioutil.TempDir("", "build-application")
		Expect(err).NotTo(HaveOccurred())

		ctx.Buildpack.Metadata = map[string]interface{}{
			"configurations": []map[string]interface{}{
				{
					"name":    "BP_APPLICATION_SCRIPT",
					"default": "*/bin/*",
				},
			},
		}
		ctx.Plan = libcnb.BuildpackPlan{Entries: []libcnb.BuildpackPlanEntry{
			{
				Name: "jvm-application",
			},
		}}
	})

	it.After(func() {
		Expect(os.RemoveAll(ctx.Application.Path)).To(Succeed())
	})

	context("DistZip exists", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(ctx.Application.Path, "app", "bin"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(ctx.Application.Path, "app", "bin", "test-script"), []byte{}, 0755))
		})

		it("contributes processes", func() {
			result, err := distzip.Build{}.Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Processes).To(ContainElements(
				libcnb.Process{Type: "dist-zip", Command: filepath.Join(ctx.Application.Path, "app", "bin", "test-script")},
				libcnb.Process{Type: "task", Command: filepath.Join(ctx.Application.Path, "app", "bin", "test-script")},
				libcnb.Process{Type: "web", Command: filepath.Join(ctx.Application.Path, "app", "bin", "test-script"), Default: true},
			))
		})

		context("$BP_LIVE_RELOAD_ENABLED is true", func() {
			it.Before(func() {
				Expect(os.Setenv("BP_LIVE_RELOAD_ENABLED", "true")).To(Succeed())
			})

			it.After(func() {
				Expect(os.Unsetenv("BP_LIVE_RELOAD_ENABLED")).To(Succeed())
			})

			it("contributes reloadable process type", func() {
				result, err := distzip.Build{}.Build(ctx)
				Expect(err).NotTo(HaveOccurred())

				Expect(result.Processes).To(ContainElements(
					libcnb.Process{Type: "dist-zip", Command: filepath.Join(ctx.Application.Path, "app", "bin", "test-script")},
					libcnb.Process{Type: "task", Command: filepath.Join(ctx.Application.Path, "app", "bin", "test-script")},
					libcnb.Process{Type: "web", Command: filepath.Join(ctx.Application.Path, "app", "bin", "test-script")},
					libcnb.Process{Type: "reload", Command: "watchexec", Arguments: []string{"-r", filepath.Join(ctx.Application.Path, "app", "bin", "test-script")}, Default: true},
				))
			})
		})
	})

	context("DistZip does not exists", func() {
		it("passes plan entries to subsequent buildpacks", func() {
			result, err := distzip.Build{}.Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Processes).To(BeEmpty())
			Expect(len(result.Unmet)).To(Equal(1))
			Expect(result.Unmet[0].Name).To(Equal("jvm-application"))
		})
	})
}
