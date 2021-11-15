// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"bytes"
	"fmt"
	"io"
	goexec "os/exec"

	"github.com/vmware-tanzu/carvel-kapp-controller/pkg/apis/kappctrl/v1alpha1"
	"github.com/vmware-tanzu/carvel-kapp-controller/pkg/exec"
	"github.com/vmware-tanzu/carvel-kapp-controller/pkg/memdir"
)

type Kbld struct {
	opts        v1alpha1.AppTemplateKbld
	genericOpts GenericOpts
}

var _ Template = &Kbld{}

func NewKbld(opts v1alpha1.AppTemplateKbld, genericOpts GenericOpts) *Kbld {
	return &Kbld{opts, genericOpts}
}

func (t *Kbld) TemplateDir(dirPath string) (exec.CmdRunResult, bool) {
	return t.template(dirPath, nil), true
}

func (t *Kbld) TemplateStream(input io.Reader, dirPath string) exec.CmdRunResult {
	return t.template(dirPath, input)
}

func (t *Kbld) template(dirPath string, input io.Reader) exec.CmdRunResult {
	args, err := t.addPaths(dirPath, input, []string{})
	if err != nil {
		return exec.NewCmdRunResultWithErr(err)
	}

	var stdoutBs, stderrBs bytes.Buffer

	cmd := goexec.Command("kbld", args...)
	cmd.Stdin = input
	cmd.Stdout = &stdoutBs
	cmd.Stderr = &stderrBs
	cmd.Dir = dirPath

	err = cmd.Run()

	result := exec.CmdRunResult{
		Stdout: stdoutBs.String(),
		Stderr: stderrBs.String(),
	}
	result.AttachErrorf("Templating dir: %s", err)

	return result
}

func (t *Kbld) addPaths(dirPath string, input io.Reader, args []string) ([]string, error) {
	// If explicit paths provided, expect user specify stdin explicitly
	switch {
	case len(t.opts.Paths) > 0:
		for _, path := range t.opts.Paths {
			if path == stdinPath {
				if input == nil {
					return nil, fmt.Errorf("Expected stdin to be available when using it as path, but was not")
				}
				args = append(args, "-f", path)
			} else {
				checkedPath, err := memdir.ScopedPath(dirPath, path)
				if err != nil {
					return nil, fmt.Errorf("Checking path: %s", err)
				}
				args = append(args, "-f", checkedPath)
			}
		}
		return args, nil

	case input != nil:
		return append(args, "-f", "-"), nil

	default:
		return append(args, "-f", dirPath), nil
	}
}
