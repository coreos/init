#!/bin/bash
# Copyright (c) 2013 The CoreOS Authors. All rights reserved.
# Use of this source code is governed by a BSD-style license that can be
# found in the LICENSE file.

if [[ $(id -u) -ne 0 ]]; then
    echo "This test script uses losetup and therefor must be run as root." >&2
    echo "Sorry, dealing with block devices is just kinda that way." >&2
    exit 1
fi

if ! type -p sgdisk &>/dev/null; then
    echo "sgdisk from the gdisk or gptfdisk package is not installed." >&2
    exit 1
fi

SCRIPT_PATH=$(readlink -f "$(dirname "$0")/../scripts/resize_state")
if [[ ! -x "${SCRIPT_PATH}" ]]; then
    echo "script doesn't exist or isn't executable: $SCRIPT_PATH" >&2
    exit 1
fi

set -e

DISK_IMAGE=$(mktemp --tmpdir test_resize_state.XXXXXXXXXX)
trap "rm -f '${DISK_IMAGE}'" EXIT

echo "# Creating an initial 50MB block device and filesystem."
truncate --size=50M "${DISK_IMAGE}"
sgdisk --largest-new=1 --change-name=1:STATE "${DISK_IMAGE}"
DISK_LOOP=$(losetup -P -f --show "${DISK_IMAGE}")
STATE_DEV="${DISK_LOOP}p1"
trap "losetup -d "${DISK_LOOP}"; rm -f '${DISK_IMAGE}'" EXIT

mkfs.ext4 "${STATE_DEV}"
sgdisk --verify "${DISK_LOOP}"
DISK_INFO=$(sgdisk --print "${DISK_LOOP}")
FS_INFO=$(dumpe2fs "${STATE_DEV}" 2>/dev/null)

echo "# First run, this should do nothing at all."
$SCRIPT_PATH "${STATE_DEV}"

echo "# Asserting that nothing has changed."
[[ $(sgdisk --print "${DISK_LOOP}") == "$DISK_INFO" ]]
[[ $(dumpe2fs "${STATE_DEV}" 2>/dev/null) == "$FS_INFO" ]]

echo "# Extending the block device to 75MB."
truncate --size=75M "${DISK_IMAGE}"
losetup -c "${DISK_LOOP}"

echo "# Second run, this should change things."
$SCRIPT_PATH "${STATE_DEV}"

echo "# Asserting that things have changed."
[[ $(sgdisk --print "${DISK_LOOP}") != "$DISK_INFO" ]]
[[ $(dumpe2fs "${STATE_DEV}" 2>/dev/null) != "$FS_INFO" ]]

echo "# Success!"
