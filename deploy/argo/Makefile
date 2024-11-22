# Replace existing version (CURRENT_VERSION var) with the provided value (NEW_VERSION var)
# Example: make update-version CURRENT_VERSION=0.0.0 NEW_VERSION=0.0.1
.PHONY: update-version
update-version:
	sed -i '' "s/$(CURRENT_VERSION)/$(NEW_VERSION)/g" ./applications/Chart.yaml

	# All matched versions in the components directory
	find ./components -type f -exec sed -i "" "s/$(CURRENT_VERSION)/$(NEW_VERSION)/g" {} +
