use cosmwasm::traits::{ReadonlyStorage, Storage};

use crate::error::GoResult;
use crate::memory::Buffer;

// this represents something passed in from the caller side of FFI
#[repr(C)]
pub struct db_t {}

// These functions should return GoResult but because we don't trust them here, we treat the return value as i32
// These functions should return GoResult but because we don't trust them here, we treat the return value as i32
// and then check it when converting to GoResult manually
#[repr(C)]
pub struct DB_vtable {
    pub read_db: extern "C" fn(*mut db_t, Buffer, *mut Buffer) -> i32,
    pub write_db: extern "C" fn(*mut db_t, Buffer, Buffer),
}

#[repr(C)]
pub struct DB {
    pub state: *mut db_t,
    pub vtable: DB_vtable,
}

impl ReadonlyStorage for DB {
    fn get(&self, key: &[u8]) -> Option<Vec<u8>> {
        let buf = Buffer::from_vec(key.to_vec());
        let mut result_buf = Buffer::default();
        let go_result: GoResult =
            (self.vtable.read_db)(self.state, buf, &mut result_buf as *mut Buffer).into();
        match go_result {
            GoResult::Ok => { /* continue */ }
            _ => {
                // TODO handle this better. in
                // This `.consume()` is safe because we initialise `buf` from a vec just a few lines above here
                panic!("Go panicked while reading key {:?}", unsafe {
                    buf.consume()
                });
            }
        }

        if result_buf.ptr.is_null() {
            return None;
        }

        // We initialize `result_buf` with a null pointer. if it is not null,
        // that means it was initialized by the go code, with values generated by `memory::allocate_rust`
        unsafe { Some(result_buf.consume()) }
    }
}

impl Storage for DB {
    fn set(&mut self, key: &[u8], value: &[u8]) {
        let buf = Buffer::from_vec(key.to_vec());
        let buf2 = Buffer::from_vec(value.to_vec());
        // caller will free input
        (self.vtable.write_db)(self.state, buf, buf2);
    }
}
