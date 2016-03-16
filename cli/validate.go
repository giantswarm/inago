package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/controller"
)

var (
	validateCmd = &cobra.Command{
		Use:   "validate [directory...]",
		Short: "Validate groups",
		Long:  "Validate group directories on the local filesystem",
		Run:   validateRun,
	}
)

func validateRun(cmd *cobra.Command, args []string) {
	groups := args

	// If no groups are specified, assume all directories in current
	// directory are groups to be checked.
	if len(args) == 0 {
		files, err := ioutil.ReadDir(".")
		if err != nil {
			newLogger.Error(newCtx, "%#v", maskAny(err))
			os.Exit(1)
		}

		for _, file := range files {
			if file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
				// If the directory is empty, don't validate it
				subfiles, err := ioutil.ReadDir(file.Name())
				if err != nil {
					newLogger.Critical(newCtx, "%#v", maskAny(err))
					os.Exit(1)
				}
				if len(subfiles) == 0 {
					continue
				}
				// The directory contains files, add it to the list of groups
				// that should be validated
				groups = append(groups, file.Name())
			}
		}
	}

	sort.Strings(groups)

	requests := []controller.Request{}
	for _, group := range groups {
		newRequestConfig := controller.DefaultRequestConfig()
		newRequestConfig.Group = group
		request := controller.NewRequest(newRequestConfig)

		request, err := extendRequestWithContent(fs, request)
		if err != nil {
			newLogger.Error(newCtx, "%#v", maskAny(err))
			os.Exit(1)
		}
		requests = append(requests, request)
	}

	for _, request := range requests {
		ok, err := controller.ValidateRequest(request)
		if ok {
			fmt.Printf("Group '%v' is valid.\n", request.Group)
		} else {
			fmt.Printf("Group '%v' not valid: %v", request.Group, err)
		}
	}

	ok, err := controller.ValidateMultipleRequest(requests)
	if ok {
		fmt.Println("Groups are valid globally.")
	} else {
		fmt.Printf("Groups are not valid globally: %v\n", err)
	}
}
