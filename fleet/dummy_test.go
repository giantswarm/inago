package fleet

import (
	"testing"

	"golang.org/x/net/context"
)

const (
	UnitName    = "unit.service"
	UnitContent = "some content"
)

// TestDummyFleet__Submit tests the DummyFleet Submit method.
func TestDummyFleet__Submit(t *testing.T) {
	dummyFleet := NewDummyFleet(DefaultDummyConfig())

	if err := dummyFleet.Submit(context.Background(), UnitName, UnitContent); err != nil {
		t.Fatal("Error submitting the test unit:", err)
	}

	statusSubmit, err := dummyFleet.GetStatus(context.Background(), UnitName)
	if err != nil {
		t.Fatal("Error getting test unit status after submit:", err)
	}

	if statusSubmit.Current != unitStateLoaded {
		t.Fatal("Incorrect current unit status after submit:", statusSubmit.Current)
	}
	if statusSubmit.Desired != unitStateLoaded {
		t.Fatal("Incorrect desired unit status after submit:", statusSubmit.Desired)
	}
	if len(statusSubmit.Machine) > 0 {
		t.Fatal("Machine status is incorrectly present")
	}
	if statusSubmit.Name != UnitName {
		t.Fatal("Incorrect unit name after submission:", statusSubmit.Name)
	}
	if statusSubmit.SliceID != "" {
		t.Fatal("Incorrect slice ID after submission:", statusSubmit.SliceID)
	}
}

// TestDummyFleet__Start tests the DummyFleet Start method.
func TestDummyFleet__Start(t *testing.T) {
	dummyFleet := NewDummyFleet(DefaultDummyConfig())

	if err := dummyFleet.Start(context.Background(), UnitName); !IsUnitNotFound(err) {
		t.Fatal("Unit not found err not returned")
	}

	dummyFleet.Submit(context.Background(), UnitName, UnitContent)

	if err := dummyFleet.Start(context.Background(), UnitName); err != nil {
		t.Fatal("Error starting the test unit:", err)
	}

	statusStart, err := dummyFleet.GetStatus(context.Background(), UnitName)
	if err != nil {
		t.Fatal("Error getting test unit status after start:", err)
	}

	if statusStart.Current != unitStateLaunched {
		t.Fatal("Incorrect current unit status after start:", statusStart.Current)
	}
	if statusStart.Desired != unitStateLaunched {
		t.Fatal("Incorrect desired unit status after start:", statusStart.Desired)
	}
}

// TestDummyFleet__Stop tests the DummyFleet Stop method.
func TestDummyFleet__Stop(t *testing.T) {
	dummyFleet := NewDummyFleet(DefaultDummyConfig())

	if err := dummyFleet.Stop(context.Background(), UnitName); !IsUnitNotFound(err) {
		t.Fatal("Unit not found err not returned")
	}

	dummyFleet.Submit(context.Background(), UnitName, UnitContent)
	dummyFleet.Start(context.Background(), UnitName)

	if err := dummyFleet.Stop(context.Background(), UnitName); err != nil {
		t.Fatal("Error stopping the test unit:", err)
	}

	statusStop, err := dummyFleet.GetStatus(context.Background(), UnitName)
	if err != nil {
		t.Fatal("Error getting test unit status after stop:", err)
	}

	if statusStop.Current != unitStateLoaded {
		t.Fatal("Incorrect current unit status after stop:", statusStop.Current)
	}
	if statusStop.Desired != unitStateLoaded {
		t.Fatal("Incorrect desired unit status after stop:", statusStop.Desired)
	}
}

// TestDummyFleet__Destroy tests the DummyFleet Destroy method.
func TestDummyFleet__Destroy(t *testing.T) {
	dummyFleet := NewDummyFleet(DefaultDummyConfig())

	if err := dummyFleet.Destroy(context.Background(), UnitName); !IsUnitNotFound(err) {
		t.Fatal("Unit not found err not returned")
	}

	dummyFleet.Submit(context.Background(), UnitName, UnitContent)
	dummyFleet.Start(context.Background(), UnitName)
	dummyFleet.Stop(context.Background(), UnitName)

	if err := dummyFleet.Destroy(context.Background(), UnitName); err != nil {
		t.Fatal("Error destroying the test unit:", err)
	}

	if _, err := dummyFleet.GetStatus(context.Background(), UnitName); !IsUnitNotFound(err) {
		t.Fatal("Unit not found err not returned")
	}
}

// TestDummyFleet__GetStatusWithMatcher tests the DummyFleet GetStatusWithMatcher method.
func TestDummyFleet__GetStatusWithMatcher(t *testing.T) {
	dummyFleet := NewDummyFleet(DefaultDummyConfig())

	if _, err := dummyFleet.GetStatusWithMatcher(
		func(s string) bool { return true },
	); !IsUnitNotFound(err) {
		t.Fatal("Unit not found err not returned")
	}

	dummyFleet.Submit(context.Background(), UnitName, UnitContent)

	submitUnitStatusList, err := dummyFleet.GetStatusWithMatcher(
		func(s string) bool { return s == UnitName },
	)
	if err != nil {
		t.Fatal("Error getting status:", err)
	}
	if len(submitUnitStatusList) != 1 && submitUnitStatusList[0].Name != UnitName {
		t.Fatal("Incorrect unit status list returned")
	}

	incorrectUnitStatusList, err := dummyFleet.GetStatusWithMatcher(
		func(s string) bool { return s != UnitName },
	)
	if err != nil {
		t.Fatal("Error getting status:", err)
	}
	if len(incorrectUnitStatusList) != 0 {
		t.Fatal("Incorrect unit status list returned")
	}

	dummyFleet.Submit(context.Background(), "another-unit.service", UnitContent)

	multipleUnitStatusList, err := dummyFleet.GetStatusWithMatcher(
		func(s string) bool { return s == UnitName },
	)
	if err != nil {
		t.Fatal("Error getting status:", err)
	}
	if len(multipleUnitStatusList) != 1 && multipleUnitStatusList[0].Name != "another-unit.service" {
		t.Fatal("Incorrect unit status list returned")
	}
}
