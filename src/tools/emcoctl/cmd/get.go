/*
Copyright © 2020 Intel Corp

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the resource(s) based on the URL",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get called")
		c := NewRestClient()
		if len(args) >= 1 {
			fmt.Println(args[0])
			c.RestClientGet(args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
