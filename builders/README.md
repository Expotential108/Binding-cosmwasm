# Cross Compilation Scripts

As this library is targetting go developers, we cannot assume a properly set up
rust environment on their system. Further, when importing this library, there is no
clean way to add a `libwasmvm.{so,dll,dylib}`. It needs to be committed with the
tagged (go) release in order to be easily usable.

The solution is to precompile the rust code into libraries for the major platforms
(Linux, Windows, MacOS) and commit them to the repository at each tagged release.
This should be doable from one host machine, but is a bit tricky. This folder
contains build scripts and a Docker image to create all dynamic libraries from one
host. In general this is set up for a Linux host, but any machine that can run Docker
can do the cross-compilation.

## Changelog

**Version 0003:**

- Avoid pre-fetching of dependences to decouple builders from source code.
- Bump `OSX_VERSION_MIN` to 10.10.
- Use `rust:1.47.0-buster` as base image for cross compilation to macOS

**Version 0002:**

- Update hardcoded library name from `libgo_cosmwasm` to `libwasmvm`.

**Version 0001:**

- First release of builders that is versioned separately of CosmWasm.
- Update Rust to nightly-2020-10-24.

## Usage

Create a local docker image, capable of cross-compling linux and macos dynamic libs:

```sh
cd builders && make docker-images
```

Then in the repo root, `make release-build` will use the above docker image and
copy the generated `{so,dylib}` files into `api` directory to be linked.

## Future Work

* Add support for cross-compiling to Windows as well.
* Publish docker images when they are stable