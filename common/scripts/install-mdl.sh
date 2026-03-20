#!/bin/bash
#
# Copyright 2026 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

set -euo pipefail

LOCALBIN="${1:?Usage: $0 <localbin-path> <mdl-version>}"
MDL_VERSION="${2:?Usage: $0 <localbin-path> <mdl-version>}"
GEM_HOME_DIR="${LOCALBIN}/.gems"
WRAPPER="${LOCALBIN}/mdl"

mkdir -p "${GEM_HOME_DIR}"

# On Ruby < 3.1 the latest mixlib-shellout (a transitive dependency of mdl) requires Ruby >= 3.1,
# so rubygems would fail resolving it. Pre-install the last compatible version explicitly.
RUBY_MINOR="$(ruby -e 'puts RUBY_VERSION.split(".")[0..1].join(".")' 2>/dev/null || echo "0.0")"
if awk "BEGIN{exit !($RUBY_MINOR < 3.1)}"; then
  echo ">>> Ruby ${RUBY_MINOR} detected — pre-installing mixlib-shellout 3.3.8 for compatibility"
  GEM_HOME="${GEM_HOME_DIR}" gem install mixlib-shellout --version 3.3.8 --no-document
fi

# Install mdl into isolated gem dir
if GEM_HOME="${GEM_HOME_DIR}" "${GEM_HOME_DIR}/bin/mdl" --version >/dev/null 2>&1; then
  echo ">>> mdl already installed in ${GEM_HOME_DIR}"
  GEM_HOME="${GEM_HOME_DIR}" "${GEM_HOME_DIR}/bin/mdl" --version
else
  echo ">>> Installing mdl ${MDL_VERSION} into ${GEM_HOME_DIR}"
  GEM_HOME="${GEM_HOME_DIR}" gem install mdl --version "${MDL_VERSION}" --no-document
fi

# Create/update wrapper script so mdl is callable directly on PATH
cat > "${WRAPPER}" <<WRAPPER_SCRIPT
#!/bin/bash
exec env GEM_HOME="${GEM_HOME_DIR}" "${GEM_HOME_DIR}/bin/mdl" "\$@"
WRAPPER_SCRIPT
chmod +x "${WRAPPER}"
echo ">>> mdl wrapper installed at ${WRAPPER}"
