System Initialization for CoreOS
================================

CoreOS uses systemd for init and process management. This repo is
divided into three collections of files:

  * configs: Customized daemon configuration files, usually installed
    into ``/etc`` but really the sky is the limit there.
  * scripts: Helper scripts for init and service startup. These are
    generally to be used as systemd oneshot services and installed
    into ``/usr/lib/coreos``.
  * systemd: Unit files for mount points, our helper scripts, or other
    services that doen't install their own unit files. There is a top
    level ``coreos-startup.target`` which depends on everything that
    should be enabled by default on CoreOS.
    See [Ordering](#ordering) for details.

The coreos-base/coreos-init ebuild handles the install process.

Important Steps
---------------

A few notes on things that must happen which are unique to Core OS.

  * resize ``state``: Support easy VM growth by checking
    if there is unused space at the end of the disk and expanding the
    filesystem to use it.
  * mount ``/media/state``: Anything that should persist
    across boots *must* be in here bind mounted or linked here
    including ``/home`` and ``/var``.
  * initialize ``/media/state``: The state partition is
    created during build but we need to be sure that it has the
    directories we expect before trying to use it.
  * bind mounts: Map things into ``/media/state`` or
    ``/run`` and so on as apropraite.
  * mount ``/usr/share/oem``: Provides extra vendor add-ons.
  * generate ssh keys: The stock sshd units do not handle this so
    we need to.
  * setup dev/debug mode: Optionally bind ``dev_image`` to
    ``/usr/local`` and remount ``/`` as read-write.

Ordering
--------

Dependencies between the units can be not so obvious to a human and our
boot involves an unusually long sequence of actions:

  1. ``resize-state.service``
      Runs ``scripts/resize_stateful_partition``
  2. ``local-fs-pre.target``
  3. ``media-state.service`` and ``usr-share-oem.mount``
      Other things are likely being mounted around here too.
  4. ``init-state.service``
      Create directories needed for home, var, and var-run.
  5. ``home.mount`` and ``var.mount``
      Almost a normal looking system now!
  6. ``var-run.mount``
      Binds ``/var/run`` to ``/run`` to consolidate clutter.
  7. ``local-fs.target``
      Mounting complete!
  8. ``systemd-tmpfiles-setup.service``
      Recreates assorted files and directories based on tmpfiles.d configs.
  9. ``sysinit.target``
      Things are looking usable now! Most daemons will start about now.

