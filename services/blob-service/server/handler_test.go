package server

import (
	"bytes"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gavv/httpexpect"
	"github.com/stretchr/testify/mock"
	"github.com/veganbase/backend/chassis"
	chassis_mocks "github.com/veganbase/backend/chassis/mocks"
	"github.com/veganbase/backend/services/blob-service/db"
	"github.com/veganbase/backend/services/blob-service/mocks"
	"github.com/veganbase/backend/services/blob-service/model"
)

func assocBlob(in *model.Blob, assocs []string) *model.Blob {
	return &model.Blob{
		ID:              in.ID,
		URI:             in.URI,
		Format:          in.Format,
		Size:            in.Size,
		Owner:           in.Owner,
		Tags:            in.Tags,
		CreatedAt:       in.CreatedAt,
		AssociatedItems: assocs,
	}
}

var (
	user1      = "usr_TESTUSER1"
	user2      = "usr_TESTUSER2"
	user3      = "usr_TESTUSER3"
	tags1      = chassis.Tags([]string{"beach", "seaside"})
	tags2      = chassis.Tags([]string{"cake"})
	createdAt1 = time.Date(2019, 6, 10, 12, 0, 0, 0, time.UTC)
	createdAt2 = time.Date(2019, 6, 9, 12, 0, 0, 0, time.UTC)
	createdAt3 = time.Date(2019, 6, 8, 12, 0, 0, 0, time.UTC)
	createdAt4 = time.Date(2019, 6, 7, 12, 0, 0, 0, time.UTC)
	blob1      = model.Blob{
		ID:        "BLOB1",
		URI:       "http://img.test.com/BLOB1.jpg",
		Format:    "jpg",
		Size:      24023,
		Owner:     &user1,
		Tags:      tags1,
		CreatedAt: createdAt1,
	}
	blob2 = model.Blob{
		ID:        "BLOB2",
		URI:       "http://img.test.com/BLOB2.jpg",
		Format:    "jpg",
		Size:      16453,
		Owner:     &user1,
		Tags:      tags2,
		CreatedAt: createdAt2,
	}
	blob3 = model.Blob{
		ID:        "BLOB3",
		URI:       "http://img.test.com/BLOB3.jpg",
		Format:    "png",
		Size:      9200,
		Owner:     &user2,
		Tags:      nil,
		CreatedAt: createdAt3,
	}
	blob4 = model.Blob{
		ID:        "BLOB4",
		URI:       "http://img.test.com/BLOB4.jpg",
		Format:    "png",
		Size:      8865,
		Owner:     nil,
		Tags:      nil,
		CreatedAt: createdAt4,
	}
	blob1d = assocBlob(&blob1, []string{})
	blob2d = assocBlob(&blob2, []string{"xyz_item0001", "xyz_item0002"})
	blob3d = assocBlob(&blob3, []string{"xyz_item0001", "xyz_item0003", "xyz_item0005"})
	blob4d = assocBlob(&blob4, []string{"xyz_item0006"})
	sess1  = map[string]string{
		"X-Auth-Method":   "session",
		"X-Auth-User-Id":  "usr_TESTUSER1",
		"X-Auth-Is-Admin": "false",
	}
	sess2 = map[string]string{
		"X-Auth-Method":   "session",
		"X-Auth-User-Id":  "usr_TESTUSER2",
		"X-Auth-Is-Admin": "false",
	}
	sess3 = map[string]string{
		"X-Auth-Method":   "session",
		"X-Auth-User-Id":  "usr_TESTUSER3",
		"X-Auth-Is-Admin": "true",
	}
	dbMock      = mocks.DB{}
	storageMock = chassis_mocks.Storage{}
)

func mockSetup() {
	dbMock.
		On("BlobsByUser", "usr_TESTUSER1", []string(nil), uint(0), uint(0)).
		Return([]model.Blob{blob1, blob2}, nil)
	dbMock.
		On("BlobsByUser", "usr_TESTUSER1", []string{"beach"}, uint(0), uint(0)).
		Return([]model.Blob{blob1}, nil)

	dbMock.On("BlobByID", "BLOB1").Return(blob1d, nil)
	dbMock.On("BlobByID", "BLOB2").Return(blob2d, nil)
	dbMock.On("BlobByID", "BLOB3").Return(blob3d, nil)
	dbMock.On("BlobByID", "BLOB4").Return(blob4d, nil)
	dbMock.On("BlobByID", mock.Anything).Return(nil, db.ErrBlobNotFound)

	dbMock.On("TagsForUser", "usr_TESTUSER1").
		Return([]string{"beach", "cake", "seaside"}, nil)
	dbMock.On("TagsForUser", mock.Anything).
		Return([]string{}, nil)

	dbMock.On("CreateBlob", mock.Anything).Return(nil)
	dbMock.On("NewBlobID", mock.Anything).Return("BLOBXYZ")

	dbMock.On("SetBlobTags", "BLOB1", mock.Anything).Return(nil)
	dbMock.On("SetBlobTags", "BLOB2", mock.Anything).Return(nil)
	dbMock.On("SetBlobTags", "BLOB3", mock.Anything).Return(nil)
	dbMock.On("SetBlobTags", "BLOB4", mock.Anything).Return(nil)
	dbMock.On("SetBlobTags", mock.Anything, mock.Anything).Return(db.ErrBlobNotFound)

	dbMock.On("DeleteBlob", "BLOB1").Return(nil)
	dbMock.On("DeleteBlob", "BLOB2").Return(nil)
	dbMock.On("DeleteBlob", "BLOB3").Return(nil)
	dbMock.On("DeleteBlob", "BLOB4").Return(nil)
	dbMock.On("DeleteBlob", mock.Anything).Return(db.ErrBlobNotFound)

	dbMock.On("ClearBlobOwner", "BLOB1").Return(false, nil)
	dbMock.On("ClearBlobOwner", "BLOB2").Return(true, nil)
	dbMock.On("ClearBlobOwner", "BLOB3").Return(true, nil)
	dbMock.On("ClearBlobOwner", "BLOB4").Return(true, nil)
	dbMock.On("ClearBlobOwner", mock.Anything).Return(false, db.ErrBlobNotFound)

	dbMock.On("AddBlobsToItem", "xyz_item0001", mock.Anything).Return(nil)
	dbMock.On("AddBlobsToItem", "xyz_item0002", mock.Anything).Return(nil)
	dbMock.On("AddBlobsToItem", "xyz_item0003", mock.Anything).Return(nil)
	dbMock.On("AddBlobsToItem", "xyz_item0004", mock.Anything).Return(nil)
	dbMock.On("AddBlobsToItem", "xyz_item0005", mock.Anything).Return(nil)
	dbMock.On("AddBlobsToItem", "xyz_item0006", mock.Anything).Return(nil)
	dbMock.On("AddBlobsToItem", mock.Anything).Return(db.ErrBlobNotFound)

	dbMock.On("RemoveBlobsFromItem", "xyz_item0001", mock.Anything).Return(nil, nil)
	dbMock.On("RemoveBlobsFromItem", "xyz_item0002", mock.Anything).Return(nil, nil)
	dbMock.On("RemoveBlobsFromItem", "xyz_item0003", mock.Anything).Return(nil, nil)
	dbMock.On("RemoveBlobsFromItem", "xyz_item0004", mock.Anything).Return(nil, nil)
	dbMock.On("RemoveBlobsFromItem", "xyz_item0005", mock.Anything).Return(nil, nil)
	dbMock.On("RemoveBlobsFromItem", "xyz_item0006", mock.Anything).Return(nil, nil)
	dbMock.On("RemoveBlobsFromItem", mock.Anything).Return(db.ErrBlobNotFound)

	dbMock.
		On("SaveEvent", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	storageMock.On("Write", mock.Anything, mock.Anything, mock.Anything).
		Return("http://media-link/X.png", int64(345), nil)
	storageMock.On("Delete", mock.Anything, mock.Anything).Return(nil)
}

func RunWithServer(t *testing.T, test func(e *httpexpect.Expect)) {
	s := &Server{imageBaseURL: "http://img.test.com"}
	s.Init("blob-service", "dev", 8090, "dev", s.routes())
	dbMock = mocks.DB{}
	s.db = &dbMock
	storageMock = chassis_mocks.Storage{}
	s.blobstore = &storageMock

	srv := httptest.NewServer(s.Srv.Handler)
	defer srv.Close()

	e := httpexpect.New(t, srv.URL)

	test(e)
}

func TestList(t *testing.T) {
	RunWithServer(t, func(e *httpexpect.Expect) {
		mockSetup()

		// Unauthenticated => not found.
		e.GET("/me/blobs").
			Expect().
			Status(http.StatusNotFound)

		// Authenticated => list own blobs with /me/blobs.
		e.GET("/me/blobs").WithHeaders(sess1).
			Expect().
			Status(http.StatusOK).
			JSON().Array().Length().Equal(2)

		// Authenticated => list own blobs with /user/{id}/blobs.
		bs := e.GET("/user/usr_TESTUSER1/blobs").WithHeaders(sess1).
			Expect().
			Status(http.StatusOK).
			JSON().Array()
		bs.Length().Equal(2)
		bs.Elements(blob1, blob2)

		// Authenticated => can't list other user's blobs.
		e.GET("/user/usr_TESTUSER2/blobs").WithHeaders(sess1).
			Expect().
			Status(http.StatusNotFound)

		// Admin + authenticated => can list other user's blobs.
		e.GET("/user/usr_TESTUSER1/blobs").WithHeaders(sess3).
			Expect().
			Status(http.StatusOK).
			JSON().Array().Length().Equal(2)

		// Tag search.
		e.GET("/me/blobs").WithHeaders(sess1).WithQuery("tags", "beach").
			Expect().
			Status(http.StatusOK).
			JSON().Array().Length().Equal(1)
	})
}

func TestDetail(t *testing.T) {
	RunWithServer(t, func(e *httpexpect.Expect) {
		mockSetup()

		// Unauthenticated => not found.
		e.GET("/blob/BLOB1").
			Expect().
			Status(http.StatusNotFound)

		// Authenticated => can see own blobs.
		e.GET("/blob/BLOB1").WithHeaders(sess1).
			Expect().
			Status(http.StatusOK)

		// Authenticated => non-existent blobs not found.
		e.GET("/blob/BLOBX").WithHeaders(sess1).
			Expect().
			Status(http.StatusNotFound)

		// Authenticated => can't see other's blobs.
		e.GET("/blob/BLOB1").WithHeaders(sess2).
			Expect().
			Status(http.StatusNotFound)

		// Admin => can see other's blobs.
		e.GET("/blob/BLOB1").WithHeaders(sess3).
			Expect().
			Status(http.StatusOK)

		// Blob return includes item associations.
		e.GET("/blob/BLOB2").WithHeaders(sess1).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("associated_items").Value("associated_items").
			Array().Length().Equal(2)
	})
}

func TestTagList(t *testing.T) {
	RunWithServer(t, func(e *httpexpect.Expect) {
		mockSetup()

		// Unauthenticated => not found.
		e.GET("/blobs/tags").
			Expect().
			Status(http.StatusNotFound)

		// Authenticated => see own tags.
		e.GET("/blobs/tags").WithHeaders(sess1).
			Expect().
			Status(http.StatusOK).
			JSON().Array().Length().Equal(3)

		// Authenticated => see own tags.
		e.GET("/blobs/tags").WithHeaders(sess2).
			Expect().
			Status(http.StatusOK).
			JSON().Array().Length().Equal(0)
	})
}

// Helper to make images for blob creation testing.
func makeImage(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.NRGBA{
				R: uint8((x + y) & 255),
				G: uint8((x + y) << 1 & 255),
				B: uint8((x + y) << 2 & 255),
				A: 255,
			})
		}
	}
	return img
}

func TestCreate(t *testing.T) {
	RunWithServer(t, func(e *httpexpect.Expect) {
		mockSetup()

		// Unauthenticated => not found.
		e.POST("/blobs").
			Expect().
			Status(http.StatusNotFound)

		// Authenticated + invalid data => bad request.
		e.POST("/blobs").WithHeaders(sess1).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().
			ContainsKey("message").ValueEqual("message", "Invalid request body")

		// Authenticated + disallowed file type => bad request.
		buf := bytes.Buffer{}
		err := gif.Encode(&buf, makeImage(200, 100), &gif.Options{NumColors: 256})
		if err != nil {
			t.Error(err)
		}
		e.POST("/blobs").WithHeaders(sess1).WithBytes(buf.Bytes()).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().
			ContainsKey("message").ValueEqual("message", "Unsupported file type")

		// Authenticated + file too large => bad request.
		buf = bytes.Buffer{}
		err = jpeg.Encode(&buf, makeImage(5000, 5000), &jpeg.Options{Quality: 100})
		if err != nil {
			t.Error(err)
		}
		e.POST("/blobs").WithHeaders(sess1).WithBytes(buf.Bytes()).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().
			ContainsKey("message").ValueEqual("message", "Request body too large (limit is 10MB)")

		// Authenticated + file OK => success.
		buf = bytes.Buffer{}
		err = png.Encode(&buf, makeImage(200, 100))
		if err != nil {
			t.Error(err)
		}
		e.POST("/blobs").WithHeaders(sess1).WithBytes(buf.Bytes()).
			Expect().
			Status(http.StatusOK).
			JSON().Object().
			ContainsKey("format").ValueEqual("format", "png").
			ContainsKey("size").ValueEqual("size", len(buf.Bytes()))
	})
}

func TestUpdate(t *testing.T) {
	RunWithServer(t, func(e *httpexpect.Expect) {
		mockSetup()

		newTags := map[string][]string{"tags": []string{"different"}}
		badPatch := []string{"not", "good"}
		disallowedPatch := map[string]string{"url": "http://not.allowed"}

		// Unauthenticated => not found.
		e.PATCH("/blob/BLOB1").WithJSON(newTags).
			Expect().
			Status(http.StatusNotFound)

		// Authenticated => illegal patches are rejected.
		e.PATCH("/blob/BLOB1").WithHeaders(sess1).WithJSON(badPatch).
			Expect().
			Status(http.StatusBadRequest)
		e.PATCH("/blob/BLOB1").WithHeaders(sess1).WithJSON(disallowedPatch).
			Expect().
			Status(http.StatusBadRequest)

		// Authenticated => can update tags on own blobs.
		e.PATCH("/blob/BLOB1").WithHeaders(sess1).WithJSON(newTags).
			Expect().
			Status(http.StatusOK)

		// Authenticated => non-existent blobs not found.
		e.PATCH("/blob/BLOBX").WithHeaders(sess1).WithJSON(newTags).
			Expect().
			Status(http.StatusNotFound)

		// Authenticated => can't patch other's blobs.
		e.PATCH("/blob/BLOB1").WithHeaders(sess2).WithJSON(newTags).
			Expect().
			Status(http.StatusNotFound)

		// Admin => can patch other's blobs.
		e.PATCH("/blob/BLOB3").WithHeaders(sess3).WithJSON(newTags).
			Expect().
			Status(http.StatusOK)
	})
}

func TestDelete(t *testing.T) {
	RunWithServer(t, func(e *httpexpect.Expect) {
		mockSetup()

		// Unauthenticated => not found.
		e.DELETE("/blob/BLOB1").
			Expect().
			Status(http.StatusNotFound)

		// Authenticated => can delete own blobs (no associations =>
		// blob really deleted).
		e.DELETE("/blob/BLOB1").WithHeaders(sess1).
			Expect().
			Status(http.StatusNoContent)

		// Authenticated => can delete own blobs (with associations =>
		// blob not really deleted).
		e.DELETE("/blob/BLOB2").WithHeaders(sess1).
			Expect().
			Status(http.StatusNoContent)

		// Authenticated => can't delete other's blobs.
		e.DELETE("/blob/BLOB3").WithHeaders(sess1).
			Expect().
			Status(http.StatusNotFound)

		// Admin => can delete other's blobs.
		e.DELETE("/blob/BLOB3").WithHeaders(sess3).
			Expect().
			Status(http.StatusNoContent)
	})
}

func TestItemAssociations(t *testing.T) {
	RunWithServer(t, func(e *httpexpect.Expect) {
		mockSetup()

		// Add new item association.
		e.POST("/blob-item-assoc/xyz_item0004").
			WithJSON([]string{"BLOB2"}).
			Expect().
			Status(http.StatusNoContent)

		// Delete existing item association.
		e.DELETE("/blob-item-assoc/xyz_item0003").
			WithJSON([]string{"BLOB3"}).
			Expect().
			Status(http.StatusNoContent)

		// Delete by item.
		e.DELETE("/blob-item-assoc/xyz_item0006").
			Expect().
			Status(http.StatusNoContent)
	})
}
