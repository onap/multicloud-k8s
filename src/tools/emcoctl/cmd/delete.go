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

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the resources from input file or url from command line",
	Run: func(cmd *cobra.Command, args []string) {
		c := NewRestClient()
		if len(inputFiles) > 0 {
			resources := readResources()
			for i := len(resources) - 1; i >= 0; i-- {
				res := resources[i]
				c.RestClientDelete(res.anchor, res.body)
			}
		} else if len(args) >= 1 {
			c.RestClientDeleteAnchor(args[0])
		} else {
			fmt.Println("Error: No args ")
		}
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().StringSliceVarP(&inputFiles, "filename", "f", []string{}, "Filename of the input file")
}
