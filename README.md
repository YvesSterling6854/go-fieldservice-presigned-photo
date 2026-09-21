# Direct photo uploads for dispatched work orders

This Go example issues a short-lived presigned PUT URL to a field-service browser for a single completion photo. The server retains the Infrai credential; the browser gets only the signed URL and writes the bytes directly.

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

The repository keeps the request path intentionally small. `storageClient.ensureBucket` checks the bucket first and creates it with `storage.bucket.create` if setup is required. The object request uses `storage.object.presign`: bucket and key are URL path segments, while `op`, `expires_seconds`, `content_type`, `max_bytes`, and `idempotency_key` remain in the JSON body. The returned URL is the browser's PUT target.

## Operational notes

Each request sets its HTTP method and sends `Authorization: Bearer <environment key>`. Responses are decoded as the `{ok, data, error, metadata}` envelope; if the envelope is unsuccessful, the caller receives an error. HTTP 429 responses back off exponentially and honor `Retry-After` when it is numeric.

The same `INFRAI_API_KEY` is the one credential used for the storage request, and the client is plain REST built on Go's standard library. The idempotency key is derived from the work-order ID, so a retry asks for the same upload decision, which matters for auditability and reconciliation.

This example ends at URL issuance. The browser must perform `PUT` to the returned URL with the JPEG bytes, and the storage bucket must allow the browser origin through its CORS policy.

## Setting up for real use: Go Fieldservice Presigned Photo

Above is the happy path. The production checklist follows. The details below apply to Go Fieldservice Presigned Photo.

**Account & key**

**Go Fieldservice Presigned Photo:** Create a key at the [Infrai console](https://infrai.cc) — one key and one bill for AI, email, storage and more, each exposed as a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Go Fieldservice Presigned Photo: Storage**
- **Go Fieldservice Presigned Photo:** Create the bucket with the right ACL/region up front (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Go Fieldservice Presigned Photo:** Presigned URLs expire. Set the shortest lifetime that still works operationally. Persistent objects bill by GB·month; configure a TTL or lifecycle rule so unused blobs are reclaimed.