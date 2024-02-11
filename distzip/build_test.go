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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/paketo-buildpacks/libpak/sbom/mocks"

	"github.com/buildpacks/libcnb"
	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"

	"github.com/paketo-buildpacks/dist-zip/v5/distzip"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect      = NewWithT(t).Expect
		sbomScanner mocks.SBOMScanner
		ctx         libcnb.BuildContext
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
		ctx.Plan = libcnb.BuildpackPlan{Entries: []libcnb.BuildpackPlanEntry{
			{
				Name: "jvm-application",
			},
		}}
		sbomScanner = mocks.SBOMScanner{}
		sbomScanner.On("ScanLaunch", ctx.Application.Path, libcnb.SyftJSON, libcnb.CycloneDXJSON).Return(nil)
	})

	context("DistZip exists", func() {
		it.Before(func() {
			Expect(os.MkdirAll(filepath.Join(ctx.Application.Path, "app", "bin"), 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(ctx.Application.Path, "app", "bin", "test-script"), []byte{}, 0755))
		})

		it("contributes processes", func() {
			result, err := distzip.Build{SBOMScanner: &sbomScanner}.Build(ctx)
			Expect(err).NotTo(HaveOccurred())

			Expect(result.Processes).To(ContainElements(
				libcnb.Process{Type: "dist-zip", Command: filepath.Join(ctx.Application.Path, "app", "bin", "test-script")},
				libcnb.Process{Type: "task", Command: filepath.Join(ctx.Application.Path, "app", "bin", "test-script")},
				libcnb.Process{Type: "web", Command: filepath.Join(ctx.Application.Path, "app", "bin", "test-script"), Default: true},
			))
			sbomScanner.AssertCalled(t, "ScanLaunch", ctx.Application.Path, libcnb.SyftJSON, libcnb.CycloneDXJSON)
		})

		context("$BP_LIVE_RELOAD_ENABLED is true", func() {
			it.Before(func() {
				t.Setenv("BP_LIVE_RELOAD_ENABLED", "true")
			})

			it("contributes reloadable process type", func() {
				result, err := distzip.Build{SBOMScanner: &sbomScanner}.Build(ctx)
				Expect(err).NotTo(HaveOccurred())

				Expect(result.Processes).To(ContainElements(
					libcnb.Process{Type: "dist-zip", Command: filepath.Join(ctx.Application.Path, "app", "bin", "test-script")},
					libcnb.Process{Type: "task", Command: filepath.Join(ctx.Application.Path, "app", "bin", "test-script")},
					libcnb.Process{Type: "web", Command: filepath.Join(ctx.Application.Path, "app", "bin", "test-script")},
					libcnb.Process{Type: "reload", Command: "watchexec", Arguments: []string{"-r", filepath.Join(ctx.Application.Path, "app", "bin", "test-script")}, Default: true},
				))
				sbomScanner.AssertCalled(t, "ScanLaunch", ctx.Application.Path, libcnb.SyftJSON, libcnb.CycloneDXJSON)
			})

			it("marks all workspace files as group read-write", func() {
				_, err := distzip.Build{SBOMScanner: &sbomScanner}.Build(ctx)
				Expect(err).NotTo(HaveOccurred())

				var modes []string
				err = filepath.Walk(ctx.Application.Path, func(path string, info fs.FileInfo, err error) error {
					if err != nil {
						return err
					}
					if path != ctx.Application.Path {
						rel, err := filepath.Rel(ctx.Application.Path, path)
						if err != nil {
							return err
						}
						modes = append(modes, fmt.Sprintf("%s %s", info.Mode(), rel))
					}

					return nil
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(modes).To(ConsistOf(
					"drwxrwxr-x app",
					"drwxrwxr-x app/bin",
					"-rwxrwxr-x app/bin/test-script",
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
