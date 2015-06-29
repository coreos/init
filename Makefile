# A simple install Makefile
DESTDIR :=

all:
	@echo "Nothing to build! Try make install"

test:
	$(MAKE) -C tests test

test-root:
	$(MAKE) -C tests test-root

install:
	install -m 755 -d \
		$(DESTDIR)/lib/udev/rules.d \
		$(DESTDIR)/usr/bin \
		$(DESTDIR)/usr/lib/coreos \
		$(DESTDIR)/usr/lib/systemd/system \
		$(DESTDIR)/usr/lib/systemd/network \
		$(DESTDIR)/usr/lib/systemd/system-generators \
		$(DESTDIR)/usr/lib/tmpfiles.d \
		$(DESTDIR)/etc/env.d \
		$(DESTDIR)/usr/share/logrotate \
		$(DESTDIR)/usr/share/ssh
	install -m 755 bin/* $(DESTDIR)/usr/bin
	install -m 755 scripts/* $(DESTDIR)/usr/lib/coreos
	install -m 644 systemd/network/* $(DESTDIR)/usr/lib/systemd/network
	install -m 755 systemd/system-generators/* \
		$(DESTDIR)/usr/lib/systemd/system-generators
	install -m 644 udev/rules.d/* $(DESTDIR)/lib/udev/rules.d
	install -m 644 configs/editor.sh $(DESTDIR)/etc/env.d/99editor
	install -m 644 configs/logrotate.conf $(DESTDIR)/usr/share/logrotate/
	install -m 600 configs/sshd_config $(DESTDIR)/usr/share/ssh/
	install -m 644 configs/ssh_config $(DESTDIR)/usr/share/ssh/
	install -m 644 configs/tmpfiles.d/* $(DESTDIR)/usr/lib/tmpfiles.d/
	cp -a systemd/system/* $(DESTDIR)/usr/lib/systemd/system
	ln -sf ../run/issue $(DESTDIR)/etc/issue

install-usr: install

.PHONY: all test test-root install-usr install
