package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/controller"
	"github.com/giantswarm/inago/validator"
)

var (
	validateCmd = &cobra.Command{
		Use:   "validate",
		Short: "validate groups",
		Long:  "validate groups",
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
			fmt.Printf("%#v\n", maskAny(err))
			os.Exit(1)
		}

		for _, file := range files {
			if file.IsDir() && !strings.HasPrefix(file.Name(), ".") {
				groups = append(groups, file.Name())
			}
		}
	}

	sort.Strings(groups)

	requests := []controller.Request{}
	for _, group := range groups {
		request, err := createRequestWithContent([]string{group})
		if err != nil {
			fmt.Printf("%#v\n", maskAny(err))
			os.Exit(1)
		}
		requests = append(requests, request)
	}

	for _, request := range requests {
		ok, err := validator.ValidateRequest(request)
		if ok {
			fmt.Printf("Group '%v' is valid.\n", request.Group)
		} else {
			fmt.Printf("Group '%v' not valid: %v.\n", request.Group, err)
		}
	}

	ok, err := validator.ValidateMultipleRequest(requests)
	if ok {
		fmt.Println("Groups are valid globally.")
	} else {
		fmt.Println("Groups are not valid globally:", err)
	}
}
