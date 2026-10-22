#!/bin/sh
# Cross-compiles tvlink into the app as libtvlink.so for 64- and 32-bit ARM TVs.
# Run from anywhere, before ./gradlew assembleRelease. GOOS=android turns on
# anet, which lists network interfaces on Android 11+; it needs -checklinkname=0.
# 32-bit ARM Android builds need cgo, so set ANDROID_NDK_HOME to an NDK (r27+).
set -eu
here=$(cd "$(dirname "$0")" && pwd)
libs="$here/app/src/main/jniLibs"
ndk_bin="${ANDROID_NDK_HOME:?set ANDROID_NDK_HOME to an Android NDK}/toolchains/llvm/prebuilt/linux-x86_64/bin"
cd "$here/.."
build() {
	mkdir -p "$libs/$1"
	GOOS=android GOARCH="$2" go build -trimpath -ldflags="-s -w -checklinkname=0" \
		-o "$libs/$1/libtvlink.so" ./cmd/tvlink
	echo "built $libs/$1/libtvlink.so"
}
CGO_ENABLED=0 build arm64-v8a arm64
CGO_ENABLED=1 GOARM=7 CC="$ndk_bin/armv7a-linux-androideabi26-clang" build armeabi-v7a arm
