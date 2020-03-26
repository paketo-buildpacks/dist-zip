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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/dist-zip/distzip"
	"github.com/sclevine/spec"
)

func testScriptResolver(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		r distzip.ScriptResolver
	)

	it.Before(func() {
		var err error

		r.ApplicationPath, err = ioutil.TempDir("", "script-resolver")
		Expect(err).NotTo(HaveOccurred())
	})

	it.After(func() {
		Expect(os.RemoveAll(r.ApplicationPath)).To(Succeed())
	})

	it("returns script", func() {
		Expect(os.MkdirAll(filepath.Join(r.ApplicationPath, "app", "bin"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(r.ApplicationPath, "app", "bin", "alpha.sh"), []byte{}, 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(r.ApplicationPath, "app", "bin", "alpha.bat"), []byte{}, 0755)).To(Succeed())

		s, ok, err := r.Resolve()
		Expect(err).NotTo(HaveOccurred())

		Expect(ok).To(BeTrue())
		Expect(s).To(Equal(filepath.Join(r.ApplicationPath, "app", "bin", "alpha.sh")))
	})

	context("$BP_APPLICATION_SCRIPT", func() {
		it.Before(func() {
			Expect(os.Setenv("BP_APPLICATION_SCRIPT", filepath.Join("bin", "*.bat"))).To(Succeed())
		})

		it.After(func() {
			Expect(os.Unsetenv("BP_APPLICATION_SCRIPT")).To(Succeed())
		})

		it("returns script from $BP_APPLICATION_SCRIPT", func() {
			Expect(os.MkdirAll(filepath.Join(r.ApplicationPath, "bin"), 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(r.ApplicationPath, "bin", "alpha.sh"), []byte{}, 0755)).To(Succeed())
			Expect(ioutil.WriteFile(filepath.Join(r.ApplicationPath, "bin", "alpha.bat"), []byte{}, 0755)).To(Succeed())

			s, ok, err := r.Resolve()
			Expect(err).NotTo(HaveOccurred())

			Expect(ok).To(BeTrue())
			Expect(s).To(Equal(filepath.Join(r.ApplicationPath, "bin", "alpha.bat")))
		})
	})

	it("returns false for no script", func() {
		_, ok, err := r.Resolve()
		Expect(err).NotTo(HaveOccurred())

		Expect(ok).To(BeFalse())
	})

	it("returns error for multiple scripts", func() {
		Expect(os.MkdirAll(filepath.Join(r.ApplicationPath, "app", "bin"), 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(r.ApplicationPath, "app", "bin", "alpha"), []byte{}, 0755)).To(Succeed())
		Expect(ioutil.WriteFile(filepath.Join(r.ApplicationPath, "app", "bin", "bravo"), []byte{}, 0755)).To(Succeed())

		_, _, err := r.Resolve()
		Expect(err).To(MatchError(fmt.Sprintf(`unable to find application script in */bin/*, candidates: [%s %s]`,
			filepath.Join(r.ApplicationPath, "app", "bin", "alpha"),
			filepath.Join(r.ApplicationPath, "app", "bin", "bravo"))))
	})

}
