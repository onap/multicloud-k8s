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

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "apply(Post) the resources from input file or url(without body) from command line",
	Run: func(cmd *cobra.Command, args []string) {
		c := NewRestClient()
		if len(inputFiles) > 0 {
			resources := readResources()
			for _, res := range resources {
				if res.file != "" {
					err := c.RestClientMultipartPost(res.anchor, res.body, res.file)
					if err != nil {
						fmt.Println("Apply: ",  res.anchor, "Error: ",err)
						return
					}
				} else {
					err := c.RestClientPost(res.anchor, res.body)
					if err != nil {
						fmt.Println("Apply: ",  res.anchor, "Error: ",err)
						return
					}
				}
			}
		} else if len(args) >= 1 {
			c.RestClientPost(args[0], []byte{})
		} else {
			fmt.Println("Error: No args ")
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.Flags().StringSliceVarP(&inputFiles, "filename", "f", []string{}, "Filename of the input file")
	//applyCmd.Flags().StringSliceVarP(&valuesFiles, "values", "v", []string{}, "Values to go with the file")
}
