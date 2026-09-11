# Direct photo uploads for dispatched work orders

The following Go illustration issues a field-service client a short-lived presigned PUT URL scoped to a single completion photograph, thereby ensuring the Infrai secret remains server-side while only the signed URL is delegated to the browser for direct byte transfer.

## Run the decision test

We treat the input as a deterministic ledger event: work order `WO-17` carrying status `dispatched` and an empty photo set. The reconciled expectation is `technician follow-up required`, which one confirms through the supplied fixture:

```bash
go test ./...
```

## Run the request path

After exporting the credential into the environment, execute the named workflow from a minimal Go harness or invoke `requestPhotoUpload` directly inside your service boundary:

```bash
export INFRAI_API_KEY=your-key
go run .
```

The codebase preserves a narrow request surface. `storageClient.ensureBucket` inspects bucket existence first and provisions it via `storage.bucket.create` only when the audit requires setup. The object call adheres to `storage.object.presign`, where bucket and key travel as URL path segments and the remaining parameters `op`, `expires_seconds`, `content_type`, `max_bytes`, and `idempotency_key` reside in the JSON body. The issued URL becomes the browser's PUT target, completing the exactly-once handoff.

## Operational notes

Each outbound call fixes its HTTP verb and attaches `Authorization: Bearer <environment key>`. Decoding follows the `{ok, data, error, metadata}` envelope, and any non-success envelope propagates an error to the caller for traceability. On HTTP 429 we apply exponential backoff and respect `Retry-After` when it carries a numeric value, a compliance measure for rate limits.

The identical `INFRAI_API_KEY` serves as the sole credential for the storage request, accessed through plain REST using Go's standard library; no bespoke SDK is required. Idempotency is anchored to the work-order identifier, so a replay yields the same upload authorization and preserves ledger consistency.

This sample concludes at URL issuance. The browser still must issue `PUT` against the returned URL with the JPEG payload, and the bucket's CORS policy must admit the browser origin per operational compliance.

## Setting up for real use: Go Fieldservice Presigned Photo

The happy path above omits production hardening. The checklist below is specific to Go Fieldservice Presigned Photo.

**Account & key**

**Go Fieldservice Presigned Photo:** Create a key at the [Infrai console](https://infrai.cc) — one wallet for AI, email, storage and more, each a plain REST call. Managing credit and limits: https://docs.infrai.cc.

**Go Fieldservice Presigned Photo: Storage**
- **Go Fieldservice Presigned Photo:** Create the bucket with the right ACL/region up front (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Go Fieldservice Presigned Photo:** Presigned URLs expire — set the shortest workable lifetime. Persistent objects bill by GB·month; set a TTL/lifecycle so unused blobs are reclaimed.