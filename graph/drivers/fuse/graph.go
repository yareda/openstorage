// +build linux,have_fuse,experimenatl

package fuse

/*
extern int start_fuse(char *, char *);
#cgo LDFLAGS: -lfuse -lulockmgr
*/
import "C"

import (
	"fmt"
	"os"
	"path"
	"syscall"

	"github.com/libopenstorage/openstorage/api"
	"github.com/libopenstorage/openstorage/graph"

	"github.com/docker/docker/daemon/graphdriver"
	"github.com/docker/docker/daemon/graphdriver/overlay"
	"github.com/docker/docker/pkg/idtools"

	log "github.com/Sirupsen/logrus"
)

const (
	Name     = "fuse"
	Type     = api.Graph
	virtPath = "/var/lib/openstorage/fuse/virtual"
	physPath = "/var/lib/openstorage/fuse/physical"
)

type Driver struct {
	// Driver is an implementation of GraphDriver. Only select methods are overridden
	graphdriver.Driver
}

func Init(home string, options []string, uidMaps, gidMaps []idtools.IDMap) (graphdriver.Driver, error) {

	log.Infof("Initializing Fuse Graph driver at home:%s and storage: %v...", home, virtPath)

	// In case it is mounted.
	syscall.Unmount(virtPath, 0)

	err := os.MkdirAll(virtPath, 0744)
	if err != nil {
		log.Fatalf("Error while creating FUSE mount path: %v", err)
	}

	err = os.MkdirAll(physPath, 0744)
	if err != nil {
		log.Fatalf("Error while creating FUSE mount path: %v", err)
	}

	cVirtPath := C.CString(virtPath)
	cPhysPath := C.CString(physPath)
	go C.start_fuse(cPhysPath, cVirtPath)

	ov, err := overlay.Init(home, options, uidMaps, gidMaps)
	if err != nil {
		volDriver.Shutdown()
		return nil, err
	}

	d := &Driver{
		Driver: ov,
	}

	return d, nil
}

func (d *Driver) String() string {
	return "openstorage-fuse"
}

// Cleanup performs necessary tasks to release resources
// held by the driver, e.g., unmounting all layered filesystems
// known to this driver.
func (d *Driver) Cleanup() error {
	syscall.Unmount(virtPath, 0)
	return nil
}

// Status returns a set of key-value pairs which give low
// level diagnostic status about this driver.
func (d *Driver) Status() [][2]string {
	return [][2]string{
		{"OpenStorage FUSE", "OK"},
	}
}

// Create creates a new, empty, filesystem layer with the
// specified id and parent and mountLabel. Parent and mountLabel may be "".
func (d *Driver) Create(id string, parent string) error {
	path := path.Join(virtPath, string(id))
	log.Infof("Creating layer %s", path)

	err := os.MkdirAll(path, 0744)
	if err != nil {
		return fmt.Errorf("Error while creating FUSE mount path %v: %v", path, err)
	}

	return nil
}

// Remove attempts to remove the filesystem layer with this id.
func (d *Driver) Remove(id string) error {
	path := path.Join(virtPath, string(id))
	log.Infof("Removing layer %s", path)

	os.RemoveAll(path)

	return nil
}

// Returns a set of key-value pairs which give low level information
// about the image/container driver is managing.
func (d *Driver) GetMetadata(id string) (map[string]string, error) {
	return nil, nil
}

// Get returns the mountpoint for the layered filesystem referred
// to by this id. You can optionally specify a mountLabel or "".
// Returns the absolute path to the mounted layered filesystem.
func (d *Driver) Get(id, mountLabel string) (string, error) {
	path := path.Join(virtPath, string(id))
	log.Infof("Getting layer %s", path)

	return path, nil
}

// Put releases the system resources for the specified id,
// e.g, unmounting layered filesystem.
func (d *Driver) Put(id string) error {
	path := path.Join(virtPath, string(id))
	log.Infof("Putting layer %s", path)

	return nil
}

// Exists returns whether a filesystem layer with the specified
// ID exists on this driver.
// All cache entries exist.
func (d *Driver) Exists(id string) bool {
	path := path.Join(virtPath, string(id))
	log.Infof("Checking if layer %s exists", path)

	_, err := os.Stat(path)

	if err == nil {
		return true
	} else {
		return false
	}
}

func (d *Driver) ApplyDiff(id string, parent string, diff archive.Reader) (size int64, err error) {
	log.Infof("Applying diff %s on %s", id, parent)
	return d.Driver.ApplyDiff(id, parent, diff)
}

func init() {
	graph.Register("fuse", Init)

	cVirtPath := C.CString(virtPath)
	cPhysPath := C.CString(physPath)
}
