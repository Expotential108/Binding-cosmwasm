package api

/*
#include "bindings.h"

// typedefs for _cgo functions (db)
typedef int64_t (*read_db_fn)(db_t *ptr, Buffer key, Buffer *val);
typedef void (*write_db_fn)(db_t *ptr, Buffer key, Buffer val);
// and api
typedef int32_t (*humanize_address_fn)(api_t*, Buffer, Buffer*);
typedef int32_t (*canonicalize_address_fn)(api_t*, Buffer, Buffer*);

// forward declarations (db)
int64_t cGet_cgo(db_t *ptr, Buffer key, Buffer *val);
void cSet_cgo(db_t *ptr, Buffer key, Buffer val);
// and api
int32_t cHumanAddress_cgo(api_t *ptr, Buffer canon, Buffer *human);
int32_t cCanonicalAddress_cgo(api_t *ptr, Buffer human, Buffer *canon);
*/
import "C"

import "unsafe"

// Note: we have to include all exports in the same file (at least since they both import bindings.h),
// or get odd cgo build errors about duplicate definitions

/****** DB ********/

type KVStore interface {
	Get(key []byte) []byte
	Set(key, value []byte)
}

var db_vtable = C.DB_vtable{
	read_db: (C.read_db_fn)(C.cGet_cgo),
	write_db: (C.write_db_fn)(C.cSet_cgo),
}

// contract: original pointer/struct referenced must live longer than C.DB struct
// since this is only used internally, we can verify the code that this is the case
func buildDB(kv KVStore) C.DB {
	return C.DB{
		state:  (*C.db_t)(unsafe.Pointer(&kv)),
		vtable: db_vtable,
	}
}

//export cGet
func cGet(ptr *C.db_t, key C.Buffer, val *C.Buffer) (ret i64) {
	// If the SDK panics, return -1 to inform the rust side something failed
	defer func() { if recover() != nil { ret = -1 } }()
	if val == nil {
		// we received an invalid pointer
		return -1
	}

	kv := *(*KVStore)(unsafe.Pointer(ptr))
	k := receiveSlice(key)
	v := kv.Get(k)
	// v will equal nil when the key is missing
	// https://github.com/cosmos/cosmos-sdk/blob/1083fa948e347135861f88e07ec76b0314296832/store/types/store.go#L174
	if v != nil {
		*val = allocateRust(v)
	}
	// else: the Buffer on the rust side is initialised as a "null" buffer,
	// so if we don't write a non-null address to it, it will understand that
	// the key it requested does not exist in the kv store

	return 0
}

//export cSet
func cSet(ptr *C.db_t, key C.Buffer, val C.Buffer) {
	kv := *(*KVStore)(unsafe.Pointer(ptr))
	k := receiveSlice(key)
	v := receiveSlice(val)
	kv.Set(k, v)
}

/***** GoAPI *******/

type HumanAddress func([]byte) (string, error)
type CanonicalAddress func(string) ([]byte, error)

type GoAPI struct {
	HumanAddress     HumanAddress
	CanonicalAddress CanonicalAddress
}

var api_vtable = C.GoApi_vtable{
	humanize_address:     (C.humanize_address_fn)(C.cHumanAddress_cgo),
	canonicalize_address: (C.canonicalize_address_fn)(C.cCanonicalAddress_cgo),
}

// contract: original pointer/struct referenced must live longer than C.GoApi struct
// since this is only used internally, we can verify the code that this is the case
func buildAPI(api *GoAPI) C.GoApi {
	return C.GoApi{
		state:  (*C.api_t)(unsafe.Pointer(api)),
		vtable: api_vtable,
	}
}

//export cHumanAddress
func cHumanAddress(ptr *C.api_t, canon C.Buffer, human *C.Buffer) (ret i32) {
	// If the SDK panics, return -1 to inform the rust side something failed
	defer func() { if recover() != nil { ret = -1 } }()
	if human == nil {
		// we received an invalid pointer
		return -1
	}

	api := (*GoAPI)(unsafe.Pointer(ptr))
	c := receiveSlice(canon)
	h, err := api.HumanAddress(c)
	if err != nil {
		return -1
	}
	*human = allocateRust([]byte(h))
	return 0
}

//export cCanonicalAddress
func cCanonicalAddress(ptr *C.api_t, human C.Buffer, canon *C.Buffer) (ret i32) {
	// If the SDK panics, return -1 to inform the rust side something failed
	defer func() { if recover() != nil { ret = -1 } }()
	if canon == nil {
		// we received an invalid pointer
		return -1
	}

	api := (*GoAPI)(unsafe.Pointer(ptr))
	h := string(receiveSlice(human))
	c, err := api.CanonicalAddress(h)
	if err != nil {
		return -1
	}
	if c != nil {
		*canon = allocateRust(c)
	}

	// If we do not set canon to a meaningful value, then the other side will interpret that as an empty result.
	return 0
}
