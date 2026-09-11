package main

import "testing"

func TestDispatchedOrderWithoutPhotoNeedsFollowUp(t *testing.T) {
	if !needsTechnicianFollowUp(workOrder{ID: "WO-17", DispatchStatus: "dispatched", PhotoCount: 0}) {
		t.Fatal("expected a follow-up when a dispatched work order has no photo")
	}
	if needsTechnicianFollowUp(workOrder{ID: "WO-17", DispatchStatus: "dispatched", PhotoCount: 1}) {
		t.Fatal("expected an existing photo to clear the follow-up decision")
	}
}
