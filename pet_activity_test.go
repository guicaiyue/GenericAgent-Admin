package main

import (
	"fmt"
	"testing"
)

func TestPetActivityStateChatEventsDriveDesktopPetActions(t *testing.T) {
	oldAction := desktopPetAction
	oldActionForTicks := desktopPetActionForTicks
	defer func() {
		desktopPetAction = oldAction
		desktopPetActionForTicks = oldActionForTicks
	}()

	var got []string
	desktopPetAction = func(action string) {
		got = append(got, action)
	}
	desktopPetActionForTicks = func(action string, ticks int) {
		got = append(got, fmt.Sprintf("%s:%02d", action, ticks))
	}

	state := newPetActivityState()
	state.handle("chat:start")
	state.handle("chat:done")
	state.handle("chat:start")
	state.handle("chat:cancel")
	state.handle("chat:start")
	state.handle("chat:error")

	want := []string{
		petActionRunning,
		petActionIdle,
		petActionReview + ":30",
		petActionRunning,
		petActionIdle,
		petActionWaiting + ":20",
		petActionRunning,
		petActionIdle,
		petActionFailed + ":24",
	}
	if len(got) != len(want) {
		t.Fatalf("actions len=%d got=%v want=%v", len(got), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("action[%d]=%q got=%v want=%v", i, got[i], got, want)
		}
	}
}

func TestPetActivityStateChatDoneKeepsRunningBaseWhenOtherDomainActive(t *testing.T) {
	oldAction := desktopPetAction
	oldActionForTicks := desktopPetActionForTicks
	defer func() {
		desktopPetAction = oldAction
		desktopPetActionForTicks = oldActionForTicks
	}()

	var got []string
	desktopPetAction = func(action string) { got = append(got, action) }
	desktopPetActionForTicks = func(action string, ticks int) {
		got = append(got, action)
	}

	state := newPetActivityState()
	state.handle("goal:start")
	state.handle("chat:start")
	state.handle("chat:done")

	want := []string{petActionRunning, petActionRunning, petActionRunning, petActionReview}
	if len(got) != len(want) {
		t.Fatalf("actions len=%d got=%v want=%v", len(got), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("action[%d]=%q got=%v want=%v", i, got[i], got, want)
		}
	}
}
