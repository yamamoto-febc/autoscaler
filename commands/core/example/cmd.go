// Copyright 2021 The sacloud Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package example

import (
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
)

var Command = &cobra.Command{
	Use:   "example [flags]...",
	Short: "show configuration example",
	RunE:  run,
}

//go:embed example.yaml
var exampleBody []byte

func run(*cobra.Command, []string) error {
	fmt.Println(string(exampleBody))
	return nil
}
