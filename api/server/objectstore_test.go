package server

import (
	"fmt"
	"testing"

	clusterclient "github.com/libopenstorage/openstorage/api/client/cluster"
	"github.com/libopenstorage/openstorage/objectstore"
	"github.com/stretchr/testify/assert"
)

func TestObjectStoreInspectSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	objInfo := &objectstore.ObjectstoreInfo{
		UUID:     "bbf89474-053b-45c1-b24f-d1dbac52638c",
		VolumeID: "328808731955060606",
		Enabled:  false,
	}
	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreInspect(objInfo.UUID).
		Return(objInfo, nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp, err := restClient.ObjectStoreInspect(objInfo.UUID)

	assert.NoError(t, err)
	assert.Equal(t, resp.UUID, objInfo.UUID)
	assert.Equal(t, resp.VolumeID, objInfo.VolumeID)
}

func TestObjectStoreInspectWithEmptyObjectstoreIDSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	objInfo := &objectstore.ObjectstoreInfo{
		UUID:     "bbf89474-053b-45c1-b24f-d1dbac52638ic",
		VolumeID: "328808731955060606",
		Enabled:  false,
	}

	objID := ""
	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreInspect(objID).
		Return(objInfo, nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp, err := restClient.ObjectStoreInspect(objID)

	assert.NoError(t, err)
	assert.Equal(t, resp.UUID, objInfo.UUID)
	assert.Equal(t, resp.VolumeID, objInfo.VolumeID)
}

func TestObjectStoreInspectFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	objID := "objtestid-1"
	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreInspect(objID).
		Return(nil, fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	_, err = restClient.ObjectStoreInspect(objID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}
func TestObjectStoreCreateSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := "testvol1"
	objInfo := &objectstore.ObjectstoreInfo{
		UUID:     "test-uuid",
		VolumeID: "test-vol-id",
		Enabled:  false,
	}
	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreCreate(name).
		Return(objInfo, nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp, err := restClient.ObjectStoreCreate(name)

	assert.NoError(t, err)
	assert.Equal(t, resp.VolumeID, objInfo.VolumeID)
}

func TestObjectStoreCreateFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	name := "testvol1"
	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreCreate(name).
		Return(nil, fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	resp, err := restClient.ObjectStoreCreate(name)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestObjectStoreUpdateSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	enable := true
	objID := "objtestid-1"
	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreUpdate(objID, enable).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.ObjectStoreUpdate(objID, enable)

	assert.NoError(t, err)
}

func TestObjectStoreUpdateWithEmptyObjectstoreIDSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	enable := true
	objID := ""
	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreUpdate(objID, enable).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.ObjectStoreUpdate(objID, enable)

	assert.NoError(t, err)
}

func TestObjectStoreUpdateFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	enable := false
	objID := "testobjid-2"
	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreUpdate(objID, enable).
		Return(fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.ObjectStoreUpdate(objID, enable)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}

func TestObjectStoreDeleteSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	objID := "objtestid-1"
	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreDelete(objID).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.ObjectStoreDelete(objID)

	assert.NoError(t, err)
}

func TestObjectStoreDeleteWithEmptyObjectstoreIDSuccess(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	objID := ""
	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreDelete(objID).
		Return(nil)

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.ObjectStoreDelete(objID)

	assert.NoError(t, err)
}
func TestObjectStoreDeleteFailed(t *testing.T) {

	// Create a new global test cluster
	ts, tc := testClusterServer(t)
	defer ts.Close()
	defer tc.Finish()

	objID := "objtestid-1"
	// mock the cluster objectstore response
	tc.MockClusterObjectStore().
		EXPECT().
		ObjectStoreDelete(objID).
		Return(fmt.Errorf("Not Implemented"))

	// create a cluster client to make the REST call
	c, err := clusterclient.NewClusterClient(ts.URL, "v1")
	assert.NoError(t, err)

	// make the REST call
	restClient := clusterclient.ClusterManager(c)
	err = restClient.ObjectStoreDelete(objID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Not Implemented")
}
