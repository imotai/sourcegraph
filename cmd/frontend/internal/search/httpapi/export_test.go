package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/store"
	"github.com/sourcegraph/sourcegraph/internal/uploadstore/mocks"
	"github.com/sourcegraph/sourcegraph/lib/iterator"
)

func TestServeSearchJobDownload(t *testing.T) {
	observationCtx := observation.TestContextTB(t)
	logger := observationCtx.Logger

	mockUploadStore := mocks.NewMockStore()
	mockUploadStore.ListFunc.SetDefaultHook(
		func(ctx context.Context, prefix string) (*iterator.Iterator[string], error) {
			return iterator.From([]string{}), nil
		})

	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	bs := basestore.NewWithHandle(db.Handle())
	s := store.New(db, observation.TestContextTB(t))
	svc := service.New(observationCtx, s, mockUploadStore)

	router := mux.NewRouter()
	router.HandleFunc("/{id}.csv", ServeSearchJobDownload(svc))

	// no job
	{
		req, err := http.NewRequest(http.MethodGet, "/99.csv", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNotFound, w.Code)
	}

	// no blobs
	{
		// create job
		userID, err := createUser(bs, "bob")
		require.NoError(t, err)
		userCtx := actor.WithActor(context.Background(), &actor.Actor{
			UID: userID,
		})
		_, err = svc.CreateSearchJob(userCtx, "foo")
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "/1.csv", nil)
		require.NoError(t, err)

		req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: userID}))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusNoContent, w.Code)
	}

	// wrong user
	{
		userID, err := createUser(bs, "alice")
		require.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, "/1.csv", nil)
		require.NoError(t, err)

		req = req.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: userID}))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusForbidden, w.Code)
	}
}

func createUser(store *basestore.Store, username string) (int32, error) {
	admin := username == "admin"
	q := sqlf.Sprintf(`INSERT INTO users(username, site_admin) VALUES(%s, %s) RETURNING id`, username, admin)
	return basestore.ScanAny[int32](store.QueryRow(context.Background(), q))
}
