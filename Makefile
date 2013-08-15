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
		$(DESTDIR)/usr/lib/tmpfiles.d \
		$(DESTDIR)/etc/ssh
	install -m 755 bin/* $(DESTDIR)/usr/bin
	install -m 755 scripts/* $(DESTDIR)/usr/lib/coreos
	install -m 644 systemd/system/* $(DESTDIR)/usr/lib/systemd/system
	install -m 644 udev/rules.d/* $(DESTDIR)/lib/udev/rules.d
	install -m 644 -T configs/tmpfiles.conf \
		$(DESTDIR)/usr/lib/tmpfiles.d/coreos-init.conf
	install -m 644 configs/ssh_config $(DESTDIR)/etc/ssh
	install -m 600 configs/sshd_config $(DESTDIR)/etc/ssh
