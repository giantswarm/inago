package cli

import (
	"io/ioutil"
	"os"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	"github.com/giantswarm/inago/controller"
	"github.com/giantswarm/inago/logging"
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
	logger := logging.GetLogger()

	groups := args

	// If no groups are specified, assume all directories in current
	// directory are groups to be checked.
	if len(args) == 0 {
		files, err := ioutil.ReadDir(".")
		if err != nil {
			logger.Error(nil, "%#v", maskAny(err))
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
			logger.Error(nil, "%#v", maskAny(err))
			os.Exit(1)
		}
		requests = append(requests, request)
	}

	for _, request := range requests {
		ok, err := controller.ValidateRequest(request)
		if ok {
			logger.Info(nil, "Group '%v' is valid.", request.Group)
		} else {
			logger.Info(nil, "Group '%v' not valid: %v.", request.Group, err)
		}
	}

	ok, err := controller.ValidateMultipleRequest(requests)
	if ok {
		logger.Info(nil, "Groups are valid globally.")
	} else {
		logger.Info(nil, "Groups are not valid globally:", err)
	}
}
