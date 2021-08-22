// Copyright 2020 VMware, Inc.
// SPDX-License-Identifier: Apache-2.0
// +build tools

package tools

import (
	"k8s.io/code-generator"
	"sigs.k8s.io/controller-tools/cmd/controller-gen"
	"k8s.io/apimachinery/pkg/apis/testapigroup/v1"
	"github.com/gogo/protobuf/proto"
	"k8s.io/code-generator/cmd/go-to-protobuf/protoc-gen-gogo"
	"github.com/gogo/protobuf/protoc-gen-gogo"
	"github.com/gogo/protobuf/protoc-gen-gofast"
	"golang.org/x/tools/cmd/goimports"
)
