package controller

import (
	"strings"
)

func DefaultNewRequest() RequestConfig {
	newConfig := RequestConfig{
		Group: "",
		Scale: 0,
	}

	return newConfig
}

type RequestConfig struct {
	// Group represents the plain group name without any slice expression.
	Group string

	// Scale represents the number random of slice IDs to create.
	Scale init
}

// Request represents a controller request. This is used to process some action
// on the controller.
type Request struct {
	RequestConfig

	// SliceIDs contains the IDs to create. IDs can be "1", "first", "whatever",
	// "5", etc..
	SliceIDs []string

	// Units represents a list of unit files that is supposed to be extended
	// using the provided slice IDs.
	Units []Unit
}

func NewRequest(config RequestConfig) Request {
	req := controller.Request{
		Group:    config.Group,
		Scale:    config.Scale,
		SliceIDs: []string{},
		Units:    []controller.Unit{},
	}

	return req
}

var unitExp = regexp.MustCompile("@.service")

// ExtendSlices extends unit files with respect to the given slice IDs. Having
// slice IDs "1" and "2" and having unit files "foo@.service" and
// "bar@.service" results in the following extended unit files.
//
// 	 foo@1.service
// 	 bar@1.service
// 	 foo@2.service
// 	 bar@2.service
//
func (r Request) ExtendSlices() (Request, error) {
	if len(r.SliceIDs) == 0 {
		return r, nil
	}
	newRequest := Request{
		SliceIDs: r.SliceIDs,
		Units:    []Unit{},
	}

	for _, sliceID := range r.SliceIDs {
		for _, unit := range r.Units {
			newUnit := unit
			newUnit.Name = unitExp.ReplaceAllString(newUnit.Name, fmt.Sprintf("@%s.service", sliceID))
			newRequest.Units = append(newRequest.Units, newUnit)
		}
	}

	return newRequest, nil
}

func (r Request) unitByName(name string) (Unit, error) {
	for _, u := range r.Units {
		if common.UnitBase(u.Name) == common.UnitBase(name) {
			return u, nil
		}
	}

	return Unit{}, maskAny(unitNotFoundError)
}

func readUnitFiles(dir string) (map[string]string, error) {
	fileInfos, err := newFileSystem.ReadDir(dir)
	if err != nil {
		return nil, maskAny(err)
	}

	unitFiles := map[string]string{}
	for _, fileInfo := range fileInfos {
		if fileInfo.IsDir() {
			continue
		}

		raw, err := newFileSystem.ReadFile(filepath.Join(dir, fileInfo.Name()))
		if err != nil {
			return nil, maskAny(err)
		}

		unitFiles[fileInfo.Name()] = string(raw)
	}

	return unitFiles, nil
}

// submit: new slice IDs, 		 unit files content
// update: existing slice IDs, unit files content
// other:  existing slice IDs

// other
func (c controller) getExistingSliceIDs(req Request) ([]string, error) {
	unitStatusList, err := c.groupStatusWithValidate(req)
	if IsUnitNotFound(err) {
		// This happenes when there is no unit, e.g. on submit. Thus we don't need
		// to check against anything. Se we do nothing and go ahead by simply
		// creating a new random ID.
	} else if err != nil {
		return nil, maskAny(err)
	}

	var sliceIDs []string
	for _, us := range unitStatusList {
		ID, err := common.SliceID(us.Name)
		if err != nil {
			return nil, maskAny(err)
		}
		seenSliceIDs = append(sliceIDs, ID)
	}

	return sliceIDs, nil
}

func (c controller) ExtendWithExistingSliceIDs(req Request) (Request, error) {
	sliceIDs, err := c.getExistingSliceIDs(req)
	if err != nil {
		return controller.Request{}, maskAny(err)
	}
	req.SliceIDs = sliceIDs

	return req, nil
}

func (c controller) ExtendWithContent(req Request) (Request, error) {
	unitFiles, err := readUnitFiles(req.Group)
	if err != nil {
		return controller.Request{}, maskAny(err)
	}
	for name, content := range unitFiles {
		req.Units = append(req.Units, controller.Unit{Name: name, Content: content})
	}

	return req, nil
}

func contains(l []string, e string) bool {
	for _, le := range l {
		if le == e {
			return true
		}
	}

	return false
}

func (c controller) ExtendWithRandomSliceIDs(req Request) (Request, error) {
	// Lookup existing slice IDs.
	unitStatusList, err := c.groupStatusWithValidate(req)
	if err != nil {
		return Request{}, maskAny(err)
	}

	// Find enough sufficient IDs.
	var newIDs []string
	for i := 0; i < req.Scale; i++ {
		for {
			newID := NewID()

			ok, err := containsUnitStatusSliceID(unitStatusList, newID)
			if err != nil {
				return Request{}, maskAny(err)
			}
			if ok {
				// We already have this ID in the group. Try again.
				continue
			}
			if contains(newIDs, newID) {
				// We already created this ID. Try again.
				continue
			}

			newIDs = append(newIDs, newID)
			break
		}
	}
	req.SliceIDs = newIDs

	return req, nil
}
