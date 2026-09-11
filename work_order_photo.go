package main

import (
	"context"
	"fmt"
	"net/url"
)

const photoBucket = "field-service-photos"

type workOrder struct {
	ID, DispatchStatus string
	PhotoCount         int
}

func needsTechnicianFollowUp(order workOrder) bool {
	return order.DispatchStatus == "dispatched" && order.PhotoCount == 0
}

func requestPhotoUpload(ctx context.Context, client *storageClient, order workOrder) (presignedUpload, error) {
	if needsTechnicianFollowUp(order) {
		return presignedUpload{}, fmt.Errorf("technician follow-up required before photo upload")
	}
	key := url.PathEscape("work-orders/" + order.ID + "/completion.jpg")
	return client.presignPhoto(ctx, photoBucket, key, "photo-"+order.ID)
}

func main() {
	client, err := newStorageClient()
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	signed, err := requestPhotoUpload(ctx, client, workOrder{ID: "WO-17", DispatchStatus: "completed", PhotoCount: 0})
	if err != nil {
		panic(err)
	}
	fmt.Println(signed.URL)
}
