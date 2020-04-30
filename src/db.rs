use crate::iterator::GoIter;
use cosmwasm_std::{generic_err, ReadonlyStorage, StdResult, Storage};

use crate::error::GoResult;
use crate::memory::Buffer;

// this represents something passed in from the caller side of FFI
#[repr(C)]
pub struct db_t {
    _private: [u8; 0],
}

// These functions should return GoResult but because we don't trust them here, we treat the return value as i32
// and then check it when converting to GoResult manually
#[repr(C)]
pub struct DB_vtable {
    pub read_db: extern "C" fn(*mut db_t, Buffer, *mut Buffer) -> i32,
    pub write_db: extern "C" fn(*mut db_t, Buffer, Buffer) -> i32,
    pub remove_db: extern "C" fn(*mut db_t, Buffer) -> i32,
    // order -> Ascending = 1, Descending = 2
    pub scan_db: extern "C" fn(*mut db_t, Buffer, Buffer, i32, *mut GoIter) -> i32,
}

#[repr(C)]
pub struct DB {
    pub state: *mut db_t,
    pub vtable: DB_vtable,
}

impl ReadonlyStorage for DB {
    fn get(&self, key: &[u8]) -> StdResult<Option<Vec<u8>>> {
        let key = Buffer::from_vec(key.to_vec());
        let mut result_buf = Buffer::default();
        let go_result: GoResult =
            (self.vtable.read_db)(self.state, key, &mut result_buf as *mut Buffer).into();
        let key = unsafe { key.consume() };
        if !go_result.is_ok() {
            return Err(generic_err(format!(
                "Go {}: reading key {:?}",
                go_result, key
            )));
        }

        if result_buf.ptr.is_null() {
            return Ok(None);
        }
        // We initialize `result_buf` with a null pointer. if it is not null,
        // that means it was initialized by the go code, with values generated by `memory::allocate_rust`
        Ok(unsafe { Some(result_buf.consume()) })
    }

    /// Allows iteration over a set of key/value pairs, either forwards or backwards.
    ///
    /// The bound `start` is inclusive and `end` is exclusive.
    ///
    /// If `start` is lexicographically greater than or equal to `end`, an empty range is described, mo matter of the order.
    fn range<'a>(
        &'a self,
        start: Option<&[u8]>,
        end: Option<&[u8]>,
        order: cosmwasm_std::Order,
    ) -> StdResult<Box<dyn Iterator<Item = StdResult<cosmwasm_std::KV>> + 'a>> {
        // returns nul pointer in Buffer in none, otherwise proper buffer
        let start = start
            .map(|s| Buffer::from_vec(s.to_vec()))
            .unwrap_or_default();
        let end = end
            .map(|e| Buffer::from_vec(e.to_vec()))
            .unwrap_or_default();
        let mut iter = GoIter::default();
        let go_result: GoResult = (self.vtable.scan_db)(
            self.state,
            start,
            end,
            order.into(),
            &mut iter as *mut GoIter,
        )
        .into();
        let _start = unsafe { start.consume() };
        let _end = unsafe { end.consume() };

        if !go_result.is_ok() {
            return Err(generic_err(format!("Go {}: creating iterator", go_result)));
        }
        Ok(Box::new(iter))
    }
}

impl Storage for DB {
    fn set(&mut self, key: &[u8], value: &[u8]) -> StdResult<()> {
        let key = Buffer::from_vec(key.to_vec());
        let value = Buffer::from_vec(value.to_vec());
        let go_result: GoResult = (self.vtable.write_db)(self.state, key, value).into();
        let key = unsafe { key.consume() };
        let _value = unsafe { value.consume() };
        if !go_result.is_ok() {
            Err(generic_err(format!(
                "Go {}: writing key {:?}",
                go_result, key
            )))
        } else {
            Ok(())
        }
    }

    fn remove(&mut self, key: &[u8]) -> StdResult<()> {
        let key = Buffer::from_vec(key.to_vec());
        let go_result: GoResult = (self.vtable.remove_db)(self.state, key).into();
        let key = unsafe { key.consume() };
        if !go_result.is_ok() {
            Err(generic_err(format!(
                "Go {}: removing key {:?}",
                go_result, key
            )))
        } else {
            Ok(())
        }
    }
}
