# Blob service

Blobs are binary data items stored in Google Cloud Storage made
accessible via a caching and rescaling proxy at img.veganbase.com (and
img-dev.veganbase.com for staging).

Users can upload blobs from the front-end and blobs can be associated
with items. Each blob thus has a dual existence, as an image in the
uploading user's image gallery, and as an image associated with zero
or more items.

## Ownership model

When blobs are initially created, they are owned by the user that
uploads them (the user ID of the uploading user is recorded in the
`owner` field in the `blobs` database table).

When a blob is assigned to an item (as a picture or some other binary
resource), this association is recorded in a separate `blob_items`
table. This has indexes that allow us to find which blobs are used for
a given item, and which items are using a given blob.

When the uploading user deletes an image from their image gallery, the
`owner` field of the corresponding entry in the `blobs` table is set
to `NULL`, marking that no user claims ownership of this blob.

When an image is removed from an item, the association between the
corresponding blob and the item is removed from the `blob_items`
table.

When a blob no longer has any entries in the `blob_items` table *and*
its `owner` field in the `blobs` table is `NULL`, then the blob is
deleted (i.e. its row in the `blobs` table is removed and its binary
data is removed from the Google Cloud Storage bucket).

TODO (BLOB-SERVICE): THINK ABOUT ARCHIVE INSTEAD OF DELETE FOR THIS
CASE.

## External API routes relating to blobs

```
GET    /blob/{id}
GET    /me/blobs?tags={tags}&page={n}&per_page={n}
GET    /user/{id}/blobs?tags={tags}&page={n}&per_page={n}
POST   /blobs
PATCH  /blob/{id}
DELETE /blob/{id}
GET    /blobs/tags
```

## Inter-service API routes relating to blobs

```
Add blob to item
Remove blob from item
Delete all unowned blobs for an item
```

## gRPC endpoints for blob service

```
rpc CreateBlob(CreateBlobRequest) returns (BlobInfo);
rpc RetrieveBlob(SingleBlobRequest) returns (BlobInfo);
rpc UpdateBlob(UpdateBlobRequest) returns (BlobInfo);
rpc DeleteBlob(SingleBlobRequest) returns (BlobInfo);
rpc RetrieveBlobs(RetrieveBlobsRequest) returns (BlobsInfo);
rpc GetTagList(GetTagListRequest) returns (TagList);
rpc AddBlobToItem(BlobItemAssocRequest) returns (Empty);
rpc RemoveBlobFromItem(BlobItemAssocRequest) returns (Empty);
rpc RemoveItemBlobs(RemoveItemBlobsRequest) returns (Empty);
```
