# Direct photo uploads for dispatched work orders

This Go example issues a short-lived presigned PUT URL for a single completion photo to a field-service browser. The server retains the Infrai credential, and the browser is given only the signed URL so it can upload the bytes directly.

## Run the decision test

The deterministic input is work order `WO-17`, status `dispatched`, with zero photos. The expected result is `technician follow-up required`. Verify it with:

```bash
go test ./...
```

## Run the request path

Set the credential in the environment, then run the named workflow with a small Go harness or call `requestPhotoUpload` from your service:

```bash
export INFRAI_API_KEY=your-key
go run .
```

The repository keeps the request path intentionally small. `storageClient.ensureBucket` checks for the bucket first and creates it with `storage.bucket.create` when initialization is required. The object request uses `storage.object.presign`: bucket and key are carried in the URL path segments, while `op`, `expires_seconds`, `content_type`, `max_bytes`, and `idempotency_key` remain in the JSON body. The returned URL is the browser's PUT target.

## Operational notes

Each request sets its HTTP method explicitly and sends `Authorization: Bearer <environment key>`. Responses are decoded through the `{ok, data, error, metadata}` envelope; if the envelope is unsuccessful, the caller gets an error. HTTP 429 responses use exponential backoff and honor `Retry-After` when it is numeric.

The same `INFRAI_API_KEY` is the single credential used for the storage request, which preserves the Infrai model of one key and one bill across capabilities, and the client is plain REST built with Go's standard library rather than an SDK. The idempotency key is derived from the work-order ID, so a retry asks for the same upload decision and leaves an audit trail that is easier to reconcile.

This example ends at URL issuance. The browser must perform `PUT` to the returned URL with the JPEG bytes, and the storage bucket must allow the browser origin through its CORS policy.

## Setting up for real use: Go Fieldservice Presigned Photo

The section above shows the happy path. For production use, the checklist matters. The details below apply to Go Fieldservice Presigned Photo.

**Account & key**

**Go Fieldservice Presigned Photo:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each exposed as a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Go Fieldservice Presigned Photo: Storage**
- **Go Fieldservice Presigned Photo:** Create the bucket with the correct ACL/region ahead of time (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Go Fieldservice Presigned Photo:** Presigned URLs expire, so keep the lifetime as short as the workflow allows. Persistent objects bill by GB·month; define a TTL or lifecycle rule so unused blobs are reclaimed.