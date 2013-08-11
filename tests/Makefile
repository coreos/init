# Simple test runner

# Safe tests to run as any user
SAFE_TESTS := $(wildcard test_*.py)

# Tests that must be run as root :-/
ROOT_TESTS := test_resize_state.sh

# Test targets mean run them, not generate them as make normally expects
.PHONY: $(SAFE_TESTS) $(ROOT_TESTS)
$(SAFE_TESTS) $(ROOT_TESTS):
	./$@

test: $(SAFE_TESTS)
	@echo "Safe tests complete!"

test-root: test $(ROOT_TESTS)
	@echo "Root tests complete!"
