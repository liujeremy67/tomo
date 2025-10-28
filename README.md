What Needs Future Work:

S3 file deletion when media/posts are deleted (currently database-only)
Optional: Media update endpoint for reordering

S3 make uploads path .env configurable

1. Batch Deletes - Add a DeleteManyFromS3(ctx context.Context, urls []string) helper that calls DeleteObjects instead of looping one-by-one. More efficient for bulk media deletes.
2. S3 Presigned Uploads	- Eventually, switch from direct file uploads to generating presigned URLs on your backend (using s3.PresignClient). Your React Native app uploads directly to S3, avoiding large payloads through your API.
3. Unit Tests - Add simple tests for ExtractS3Key, and mock DeleteFromS3 with AWS SDK’s smithy mocks to ensure safe deletion logic.
4. Logging & Metrics - Add log.Printf or structured logs in the S3 utils so you can debug upload/delete operations easily during development.

JWT REFRESHING