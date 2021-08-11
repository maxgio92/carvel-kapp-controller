// Copyright 2021 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package secretgencontroller

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vmware-tanzu/carvel-kapp-controller/test/e2e"
)

func Test_PlaceholderSecrets_DeletedWhenPackageInstallDeleted(t *testing.T) {
	env := e2e.BuildEnv(t)
	logger := e2e.Logger{}
	kapp := e2e.Kapp{t, env.Namespace, logger}
	kubectl := e2e.Kubectl{t, env.Namespace, logger}
	name := "placeholder-garbage-collection"
	sas := e2e.ServiceAccounts{env.Namespace}

	pkgiYaml := fmt.Sprintf(`---
apiVersion: data.packaging.carvel.dev/v1alpha1
kind: PackageMetadata
metadata:
  name: pkg.test.carvel.dev
  namespace: %[1]s
spec:
  # This is the name we want to reference in resources such as PackageInstall.
  displayName: "Test PackageMetadata in repo"
  shortDescription: "PackageMetadata used for testing"
  longDescription: "A longer, more detailed description of what the package contains and what it is for"
  providerName: Carvel
  maintainers:
  - name: carvel
  categories:
  - testing
  supportDescription: "Description of support provided for the package"
---
apiVersion: data.packaging.carvel.dev/v1alpha1
kind: Package
metadata:
  name: pkg.test.carvel.dev.1.0.0
  namespace: %[1]s
spec:
  refName: pkg.test.carvel.dev
  version: 1.0.0
  licenses:
  - Apache 2.0
  capactiyRequirementsDescription: "cpu: 1,RAM: 2, Disk: 3"
  releaseNotes: |
    - Introduce simple-app package
  releasedAt: 2021-05-05T18:57:06Z
  template:
    spec:
      fetch:
      - imgpkgBundle:
          image: k8slt/kctrl-example-pkg:v1.0.0
      template:
      - ytt: {}
      - kbld:
          paths:
          - "-"
          - ".imgpkg/images.yml"
      deploy:
      - kapp: {}
---
apiVersion: packaging.carvel.dev/v1alpha1
kind: PackageInstall
metadata:
  name: %[2]s
  namespace: %[1]s
  annotations:
    kapp.k14s.io/change-group: kappctrl-e2e.k14s.io/packageinstalls
spec:
  serviceAccountName: kappctrl-e2e-ns-sa
  packageRef:
    refName: pkg.test.carvel.dev
    versionSelection:
      constraints: 1.0.0
  values:
  - secretRef:
      name: pkg-demo-values
---
apiVersion: v1
kind: Secret
metadata:
  name: pkg-demo-values
stringData:
  values.yml: |
    hello_msg: "hi"
`, env.Namespace, name) + sas.ForNamespaceYAML()

	cleanUp := func() {
		// Delete App with kubectl first since kapp
		// deletes ServiceAccount before App
		kubectl.RunWithOpts([]string{"delete", "apps/" + name}, e2e.RunOpts{AllowError: true})
		kapp.Run([]string{"delete", "-a", name})
	}
	cleanUp()
	defer cleanUp()

	logger.Section("Create PackageInstall", func() {
		kapp.RunWithOpts([]string{"deploy", "-a", name, "-f", "-"},
			e2e.RunOpts{StdinReader: strings.NewReader(pkgiYaml)})

		kubectl.Run([]string{"wait", "--for=condition=ReconcileSucceeded", "pkgi/" + name, "--timeout", "1m"})
	})

	logger.Section("Check placeholder secret created", func() {
		kubectl.Run([]string{"get", "secret", name + "-fetch0"})
	})

	logger.Section("Check placeholder secret deleted after PackageInstall deleted", func() {
		cleanUp()
		out, err := kubectl.RunWithOpts([]string{"get", "secret", name + "fetch0"}, e2e.RunOpts{AllowError: true})
		assert.NotNil(t, err, "expected error from not finding placeholder secret.\nGot: "+out)
	})
}
