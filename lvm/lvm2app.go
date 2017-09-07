package lvm

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"unsafe"
)

/*
#cgo LDFLAGS: -llvm2app -ldevmapper
#cgo CFLAGS: -Wno-implicit-function-declaration
#cgo pkg-config: --libs glib-2.0
#include <stdlib.h>
#include <lvm2app.h>
#include <libdevmapper.h>

// Returns an array of strings obtained from a `dm_list` with entries of type `lvm_str_list`.
// This is useful for obtaining a list of strings from eg., `lvm_list_vg_names()`.
char** csilvm_get_strings_from_lvm_str_list(const struct dm_list *list)
{
	struct lvm_str_list *strl;
	char **results;
	unsigned int list_size;

	list_size = dm_list_size(list);
	results = malloc(sizeof (char *) * list_size);
	int ii = 0;
	dm_list_iterate_items(strl, list) {
		results[ii] = strdup(strl->str);
		ii++;
	}
	return results;
}

*/
import "C"

// LibraryGetVersion corresponds to `lvm_library_get_version` in `lvm2app.h`.
func LibraryGetVersion() string {
	return C.GoString(C.lvm_library_get_version())
}

// LibHandle holds a handle to the `lvm2app` library. The caller must
// call `Close()` when completely done with the library.
type LibHandle struct {
	lk     sync.Mutex
	handle C.lvm_t
}

// NewLibHandle returns a new lvm2app library handle. It corresponds to
// `lvm_init`. The caller must call `Close()` on the *LibHandle that
// is returned to free the allocated memory.
func NewLibHandle() (*LibHandle, error) {
	handle := C.lvm_init(nil)
	if handle == nil {
		return nil, errors.New("lvm2app: Init: failed to initialize handle")
	}
	return &LibHandle{sync.Mutex{}, handle}, nil
}

// Err returns the error returned for the last operation, if any. The
// return type is `*Error`.
func (h *LibHandle) err() error {
	errno := C.lvm_errno(h.handle)
	if errno == 0 {
		return nil
	}
	errmsg := C.GoString(C.lvm_errmsg(h.handle))
	// Concatenate multi-line errors.
	errmsg = strings.Replace(errmsg, "\n", " ", -1)
	// Get the name of the calling function.
	pc, _, _, ok := runtime.Caller(1)
	details := runtime.FuncForPC(pc)
	caller := ""
	if ok && details != nil {
		tokens := strings.Split(details.Name(), ".")
		caller = tokens[len(tokens)-1]
	}
	return &Error{caller, errmsg, int(errno)}
}

type Error struct {
	Caller string
	Errmsg string
	Errno  int
}

func (e *Error) Error() string {
	caller := e.Caller + ": "
	return fmt.Sprintf("lvm: %s%s (%d)", caller, e.Errmsg, e.Errno)
}

// goStrings converts an array of C strings to a slice of go strings.
// See https://stackoverflow.com/questions/36188649/cgo-char-to-slice-string
func goStrings(argc C.uint, argv **C.char) []string {
	length := int(argc)
	tmpslice := (*[1 << 30]*C.char)(unsafe.Pointer(argv))[:length:length]
	gostrings := make([]string, length)
	for i, s := range tmpslice {
		gostrings[i] = C.GoString(s)
	}
	return gostrings
}

// ListVolumeGroupNames returns the names of the list of volume groups.
//
// It is equivalent to `lvm_list_vg_names` followed by
// `dm_list_iterate_items` to accumulate the string values.
//
// This does not normally scan for devices. To scan for devices, use
// the `Scan()` function.
func (handle *LibHandle) ListVolumeGroupNames() ([]string, error) {
	handle.lk.Lock()
	defer handle.lk.Unlock()
	// Get the list of volume group names
	// The memory for the `dm_list` is freed when the handle is freed.
	dm_list_names := C.lvm_list_vg_names(handle.handle)
	if dm_list_names == nil {
		// We failed to allocate memory for the list.
		return nil, handle.err()
	}
	// If the list is empty, return nil.
	// See https://github.com/twitter/bittern/blob/a95aab6d4a43c7961d36bacd9f4e23387a4cb9d7/lvm2/libdm/datastruct/list.c#L81
	if C.dm_list_empty(dm_list_names) != 0 {
		return nil, nil
	}
	size := C.dm_list_size(dm_list_names)
	if int(size) == 0 {
		// We just checked that the lists is non-empty so we
		// expect it's size to be greater than zero.
		panic("lvm2app: unexpected zero-length list")
	}
	cvgnames := C.csilvm_get_strings_from_lvm_str_list(dm_list_names)
	defer C.free(unsafe.Pointer(cvgnames))
	// Transform the array of C strings into a []string.
	vgnames := goStrings(size, cvgnames)
	return vgnames, nil
}

// ListVolumeGroupUUIDs returns the UUIDs of the list of volume groups.
//
// It is equivalent to `lvm_list_vg_names` followed by
// `dm_list_iterate_items` to accumulate the string values.
//
// This does not normally scan for devices. To scan for devices, use
// the `Scan()` function.
func (handle *LibHandle) ListVolumeGroupUUIDs() ([]string, error) {
	handle.lk.Lock()
	defer handle.lk.Unlock()
	// Get the list of volume group UUIDs
	// The memory for the `dm_list` is freed when the handle is freed.
	dm_list_uuids := C.lvm_list_vg_uuids(handle.handle)
	if dm_list_uuids == nil {
		// We failed to allocate memory for the list.
		return nil, handle.err()
	}
	// If the list is empty, return nil.
	// See https://github.com/twitter/bittern/blob/a95aab6d4a43c7961d36bacd9f4e23387a4cb9d7/lvm2/libdm/datastruct/list.c#L81
	if C.dm_list_empty(dm_list_uuids) != 0 {
		return nil, nil
	}
	size := C.dm_list_size(dm_list_uuids)
	if int(size) == 0 {
		// We just checked that the lists is non-empty so we
		// expect it's size to be greater than zero.
		panic("lvm2app: unexpected zero-length list")
	}
	cvguuids := C.csilvm_get_strings_from_lvm_str_list(dm_list_uuids)
	defer C.free(unsafe.Pointer(cvguuids))
	// Transform the array of C strings into a []string.
	vguuids := goStrings(size, cvguuids)
	return vguuids, nil
}

// LookupVolumeGroup returns the volume group with the given name.
func (handle *LibHandle) LookupVolumeGroup(name string) (*VolumeGroup, error) {
	handle.lk.Lock()
	defer handle.lk.Unlock()
	vg := &VolumeGroup{name, nil, handle}
	// Check that the volume group can be opened.
	if err := vg.open(openReadWrite); err != nil {
		return nil, err
	}
	// Close the volume group, releasing the VG lock. Subsequent
	// operations will re-open the volume group as necessary.
	vg.close()
	return vg, nil
}

// CreateVolumeGroup creates a new volume group.
func (handle *LibHandle) CreateVolumeGroup(name string, pvs []*PhysicalVolume, opts ...VolumeGroupOpt) (*VolumeGroup, error) {
	handle.lk.Lock()
	defer handle.lk.Unlock()
	// Validate the volume group name.
	if err := handle.validateVolumeGroupName(name); err != nil {
		return nil, err
	}

	// Create the volume group memory object.
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	cvg := C.lvm_vg_create(handle.handle, cname)
	if cvg == nil {
		return nil, handle.err()
	}
	vg := &VolumeGroup{name, cvg, handle}
	defer vg.close()

	// Add physical volumes to the volume group.
	for _, pv := range pvs {
		cpv := C.CString(pv.dev)
		defer C.free(unsafe.Pointer(cpv))
		res := C.lvm_vg_extend(cvg, cpv)
		if res != 0 {
			return nil, handle.err()
		}
	}

	// Persist the volume group to disk.
	res := C.lvm_vg_write(vg.vg)
	if res != 0 {
		return nil, vg.libHandle.err()
	}
	return vg, nil
}

func (handle *LibHandle) validateVolumeGroupName(name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	res := C.lvm_vg_name_validate(handle.handle, cname)
	if res != 0 {
		return handle.err()
	}
	return nil
}

type VolumeGroupOpt func(*VolumeGroup) error

// Scan scans for new devices and volume groups.
func (handle *LibHandle) Scan() error {
	handle.lk.Lock()
	defer handle.lk.Unlock()
	res := C.lvm_scan(handle.handle)
	if res != 0 {
		return handle.err()
	}
	return nil
}

// LookupPhysicalVolume returns the physical volume with the given name.
func (handle *LibHandle) LookupPhysicalVolume(name string) (*PhysicalVolume, error) {
	handle.lk.Lock()
	defer handle.lk.Unlock()
	// TODO(gpaul): confirm that the physical volume exists.
	return &PhysicalVolume{name, handle}, nil
}

// CreatePhysicalVolume creates a physical volume of the given device
// and size.
//
// If sizeInBytes is 0 the entire available space is allocated.
func (handle *LibHandle) CreatePhysicalVolume(dev string, sizeInBytes int) (*PhysicalVolume, error) {
	handle.lk.Lock()
	defer handle.lk.Unlock()
	if sizeInBytes < 0 {
		return nil, errors.New("lvm: sizeInBytes cannot be negative")
	}
	cdev := C.CString(dev)
	defer C.free(unsafe.Pointer(cdev))
	res := C.lvm_pv_create(handle.handle, cdev, C.uint64_t(sizeInBytes))
	if res != 0 {
		return nil, handle.err()
	}
	return &PhysicalVolume{dev, handle}, nil
}

// Close releases the underlying handle by calling `lvm_quit`.
func (h *LibHandle) Close() error {
	C.lvm_quit(h.handle)
	return nil
}

type PhysicalVolume struct {
	dev       string
	libHandle *LibHandle
}

// Remove removes the physical volume.
func (pv *PhysicalVolume) Remove() error {
	pv.libHandle.lk.Lock()
	defer pv.libHandle.lk.Unlock()
	cdev := C.CString(pv.dev)
	defer C.free(unsafe.Pointer(cdev))
	res := C.lvm_pv_remove(pv.libHandle.handle, cdev)
	if res != 0 {
		return pv.libHandle.err()
	}
	return nil
}

type VolumeGroup struct {
	name      string
	vg        C.vg_t
	libHandle *LibHandle
}

// BytesTotal returns the current size in bytes of the volume group.
func (vg *VolumeGroup) BytesTotal() (int, error) {
	vg.libHandle.lk.Lock()
	defer vg.libHandle.lk.Unlock()
	if err := vg.open(openReadWrite); err != nil {
		return 0, err
	}
	defer vg.close()
	return int(C.lvm_vg_get_size(vg.vg)), nil
}

// BytesFree returns the unallocated space in bytes of the volume group.
func (vg *VolumeGroup) BytesFree() (int, error) {
	vg.libHandle.lk.Lock()
	defer vg.libHandle.lk.Unlock()
	if err := vg.open(openReadWrite); err != nil {
		return 0, err
	}
	defer vg.close()
	return int(C.lvm_vg_get_free_size(vg.vg)), nil
}

var NoSpaceErr = errors.New("lvm: not enough free space")

// CreateLogicalVolume creates a logical volume of the given device
// and size.
//
// The actual size may be larger than asked for as the smallest
// increment is the size of an extent on the volume group in question.
//
// If sizeInBytes is zero the entire available space is allocated.
func (vg *VolumeGroup) CreateLogicalVolume(name string, sizeInBytes int) (*LogicalVolume, error) {
	vg.libHandle.lk.Lock()
	defer vg.libHandle.lk.Unlock()
	if sizeInBytes < 0 {
		return nil, errors.New("lvm: sizeInBytes cannot be negative")
	}
	// TODO(gpaul): validate the name with C.int lvm_lv_name_validate(const vg_t vg, const char *lv_name);
	if err := vg.open(openReadWrite); err != nil {
		return nil, err
	}
	defer vg.close()
	freeExtents := int(C.lvm_vg_get_free_extent_count(vg.vg))
	extentSize := int(C.lvm_vg_get_extent_size(vg.vg))
	if sizeInBytes == 0 {
		sizeInBytes = extentSize * freeExtents
	}
	// Calculate the number of extents required to satisfy size.
	extentsForSize := (sizeInBytes / extentSize)
	if sizeInBytes%extentSize != 0 {
		extentsForSize++
	}
	// Check that there's enough free space available.
	if extentsForSize > freeExtents {
		return nil, NoSpaceErr
	}
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	// The documentation insists that this function expects the
	// size in number of extents, but the source code appears to
	// disagree and calculates the number of extents required.
	//
	// See https://github.com/twitter/bittern/blob/master/lvm2/liblvm/lvm_lv.c#L244
	lv := C.lvm_vg_create_lv_linear(vg.vg, cname, C.uint64_t(sizeInBytes))
	if lv == nil {
		return nil, vg.libHandle.err()
	}
	actualSize := extentsForSize * extentSize
	return &LogicalVolume{name, lv, vg, actualSize}, nil
}

// LookupLogicalVolume looks up the logical volume in the volume group
// with the given name.
func (vg *VolumeGroup) LookupLogicalVolume(name string) (*LogicalVolume, error) {
	vg.libHandle.lk.Lock()
	defer vg.libHandle.lk.Unlock()
	if err := vg.open(openReadWrite); err != nil {
		return nil, err
	}
	defer vg.close()
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	lv := C.lvm_lv_from_name(vg.vg, cname)
	if lv == nil {
		return nil, vg.libHandle.err()
	}
	actualSize := int(C.lvm_lv_get_size(lv))
	return &LogicalVolume{name, lv, vg, actualSize}, nil
}

// Remove removes the volume group from disk.
//
// It calls `lvm_vg_remove` followed by `lvm_vg_write` to persist the
// change.
func (vg *VolumeGroup) Remove() error {
	vg.libHandle.lk.Lock()
	defer vg.libHandle.lk.Unlock()
	if err := vg.open(openReadWrite); err != nil {
		return err
	}
	defer vg.close()
	res := C.lvm_vg_remove(vg.vg)
	if res != 0 {
		return vg.libHandle.err()
	}
	res = C.lvm_vg_write(vg.vg)
	if res != 0 {
		return vg.libHandle.err()
	}
	return nil
}

const (
	openReadWrite = false
	openReadOnly  = true
)

// open calls `lvm_vg_open()` to get a handle to the underlying volume group.
func (vg *VolumeGroup) open(readonly bool) error {
	if vg.vg != nil {
		return errors.New("already open")
	}
	cname := C.CString(vg.name)
	defer C.free(unsafe.Pointer(cname))
	mode := "w"
	if readonly {
		mode = "r"
	}
	cmode := C.CString(mode)
	defer C.free(unsafe.Pointer(cmode))
	const ignoredFlags = 0
	cvg := C.lvm_vg_open(vg.libHandle.handle, cname, cmode, ignoredFlags)
	if cvg == nil {
		return vg.libHandle.err()
	}
	vg.vg = cvg
	return nil
}

// close calls `lvm_vg_close()` to release the underlying volume group
// handle and the VG lock. This function panics if the close fails or
// if the volume group is already closed.  As this is an internal
// function, the latter should never happen.
func (vg *VolumeGroup) close() {
	if vg.vg == nil {
		panic("already closed")
	}
	res := C.lvm_vg_close(vg.vg)
	if res != 0 {
		panic(vg.libHandle.err())
	}
	vg.vg = nil
}

func (vg *VolumeGroup) withOpen(fn func(C.vg_t) error) error {
	if err := vg.open(openReadWrite); err != nil {
		return err
	}
	defer vg.close()
	return fn(vg.vg)
}

type LogicalVolume struct {
	name        string
	lv          C.lv_t
	vg          *VolumeGroup
	sizeInBytes int
}

func (lv *LogicalVolume) Remove() error {
	lv.vg.libHandle.lk.Lock()
	defer lv.vg.libHandle.lk.Unlock()
	if err := lv.vg.open(openReadWrite); err != nil {
		return err
	}
	defer lv.vg.close()
	cvg := lv.vg.vg
	// Load the C.lv_t from scratch as the original vg_t
	// was closed and re-opened here.
	cname := C.CString(lv.name)
	defer C.free(unsafe.Pointer(cname))
	clv := C.lvm_lv_from_name(cvg, cname)
	if clv == nil {
		return lv.vg.libHandle.err()
	}
	res := C.lvm_vg_remove_lv(clv)
	if res != 0 {
		return lv.vg.libHandle.err()
	}
	return nil
}

var DefaultHandle *LibHandle

func init() {
	var err error
	DefaultHandle, err = NewLibHandle()
	if err != nil {
		panic(fmt.Errorf("lvm: init: cannot allocate lvm handle"))
	}
}

// Scan scans for new devices and volume groups.
func Scan() error {
	return DefaultHandle.Scan()
}

// LookupPhysicalVolume returns the volume group with the given name.
func LookupVolumeGroup(name string) (*VolumeGroup, error) {
	return DefaultHandle.LookupVolumeGroup(name)
}

// LookupVolumeGroup returns the volume group with the given name.
func LookupPhysicalVolume(name string) (*PhysicalVolume, error) {
	return DefaultHandle.LookupPhysicalVolume(name)
}

// CreateVolumeGroup creates a new volume group.
func CreateVolumeGroup(
	name string,
	pvs []*PhysicalVolume,
	opts ...VolumeGroupOpt) (*VolumeGroup, error) {
	return DefaultHandle.CreateVolumeGroup(name, pvs, opts...)
}

// ListVolumeGroupNames returns the names of the list of volume groups.
//
// It is equivalent to `lvm_list_vg_names` followed by
// `dm_list_iterate_items` to accumulate the string values.
//
// This does not normally scan for devices. To scan for devices, use
// the `Scan()` function.
func ListVolumeGroupNames() ([]string, error) {
	return DefaultHandle.ListVolumeGroupNames()
}

// ListVolumeGroupUUIDs returns the UUIDs of the list of volume groups.
//
// It is equivalent to `lvm_list_vg_names` followed by
// `dm_list_iterate_items` to accumulate the string values.
//
// This does not normally scan for devices. To scan for devices, use
// the `Scan()` function.
func ListVolumeGroupUUIDs() ([]string, error) {
	return DefaultHandle.ListVolumeGroupUUIDs()
}

// CreatePhysicalVolume creates a physical volume of the given device
// and size.
//
// If sizeInBytes is 0 the entire available space is allocated.
func CreatePhysicalVolume(dev string, sizeInBytes int) (*PhysicalVolume, error) {
	return DefaultHandle.CreatePhysicalVolume(dev, sizeInBytes)
}

// Extent sizing for linear logical volumes:
// https://github.com/Jajcus/lvm2/blob/266d6564d7a72fcff5b25367b7a95424ccf8089e/lib/metadata/metadata.c#L983
