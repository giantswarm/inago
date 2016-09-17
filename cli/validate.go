package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"

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

	requests := []controller.Request{}

	// If no groups are specified, assume all directories in current
	// directory are groups to be checked.
	if len(groups) == 0 {
		files, err := ioutil.ReadDir(".")
		if err != nil {
			newLogger.Error(newCtx, "%#v", maskAny(err))
			os.Exit(1)
		}
		sort.Sort(fileInfoSlice(files))
		for _, f := range files {
			if r, err := newRequestWithUnits(f.Name()); err == nil {
				requests = append(requests, r)
			}
		}
	}

	sort.Strings(groups)
	for _, group := range groups {
		request, err := newRequestWithUnits(group)
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
			validationErr := err.(controller.ValidationError)
			fmt.Printf("Group '%v' not valid: %v", request.Group, FormatValidationError(validationErr))
		}
	}

	ok, err := controller.ValidateMultipleRequest(requests)
	if ok {
		fmt.Println("Groups are valid globally.")
	} else {
		validationErr := err.(controller.ValidationError)
		fmt.Printf("Groups are not valid globally: %v\n", FormatValidationError(validationErr))
	}
}

type fileInfoSlice []os.FileInfo

func (p fileInfoSlice) Len() int           { return len(p) }
func (p fileInfoSlice) Less(i, j int) bool { return p[i].Name() < p[j].Name() }
func (p fileInfoSlice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
