# Copyright 2026 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# conditionalize-helm-rbac.awk (ILS-2352)
#
# Post-processes generated RBAC YAML, wrapping designated rule blocks and
# resource list items in Helm guards driven by a declarative gating table.
# See common/scripts/conditionalize-helm-rbac.sh and the design/plan docs.
#
# Invoked as:  awk -v TABLE=<table> -v FILEBN=<basename> -f this.awk <file>
#
# Idempotency: the script first strips any guard lines it would have added
# (openers contain `include "ibm-licensing.`; closers are indented `{{- end }}`),
# then re-applies them deterministically. Re-runs and --check are therefore
# no-ops on already-guarded files.

function trim(s) { sub(/^[ \t]+/, "", s); sub(/[ \t]+$/, "", s); return s }

function indent_of(s,   m) { m = s; sub(/[^ ].*$/, "", m); return m }

function api_in_rule(a) { return (a in rule_api) }

# Exact match, or prefix match for grouped subresources (e.g. operandrequests
# also covers operandrequests/finalizers, operandrequests/status).
function res_in_rule(rspec,   res) {
  if (rspec in rule_res) return 1
  for (res in rule_res) {
    if (index(res, rspec "/") == 1) return 1
  }
  return 0
}

# True when the rule's resource set is exactly the comma-separated set in csv.
function resset_equals(csv,   k, n, parts, want, cntWant, cntHave, x, p) {
  n = split(csv, parts, ",")
  split("", want)
  cntWant = 0
  for (k = 1; k <= n; k++) { p = trim(parts[k]); if (p != "") { want[p] = 1; cntWant++ } }
  cntHave = 0
  for (x in rule_res) { cntHave++; if (!(x in want)) return 0 }
  return (cntHave == cntWant)
}

# Build an `{{- if or (...) (...) }}` line from an "or:helperA,helperB" spec.
function build_or(spec,   parts, n, i, out, p) {
  sub(/^or:/, "", spec)
  n = split(spec, parts, ",")
  out = ""
  for (i = 1; i <= n; i++) {
    p = trim(parts[i])
    out = out " (eq (include \"" helperPrefix p "\" .) \"true\")"
  }
  return "{{- if or" out " }}"
}

BEGIN {
  helperPrefix = "ibm-licensing."
  MODE_WF = 0
  nrows = 0
  while ((getline tl < TABLE) > 0) {
    if (tl ~ /^[ \t]*#/) continue
    if (tl ~ /^[ \t]*$/) continue
    nf = split(tl, a, /\|/)
    if (nf < 7) continue
    tfile = trim(a[1])
    if (tfile != FILEBN && tfile != "*") continue
    if (trim(a[6]) == "whole-file") { MODE_WF = 1; WF_HELPER = trim(a[7]); continue }
    nrows++
    R_kind[nrows]   = trim(a[2])
    R_name[nrows]   = trim(a[3])
    R_api[nrows]    = trim(a[4])
    R_res[nrows]    = trim(a[5])
    R_action[nrows] = trim(a[6])
    R_helper[nrows] = trim(a[7])
  }
  close(TABLE)
}

{ L[++N] = $0 }

END {
  if (MODE_WF) { do_wholefile(); exit }

  # Strip any previously-injected guards so re-application is idempotent.
  cn = 0
  for (i = 1; i <= N; i++) {
    s = L[i]
    if (s ~ /include "ibm-licensing\./) continue
    if (s ~ /^[ \t]+{{- end }}[ \t]*$/) continue
    C[++cn] = s
  }

  kind = ""; name = ""; in_meta = 0; in_rules = 0; sub_state = ""; rulen = 0
  for (i = 1; i <= cn; i++) process(C[i])
  flush_rule()
}

function process(line,   item) {
  if (line ~ /^---/) {
    flush_rule()
    print line
    kind = ""; name = ""; in_meta = 0; in_rules = 0; sub_state = ""
    return
  }
  if (line ~ /^[A-Za-z]/) {
    flush_rule()
    in_rules = 0; sub_state = ""
    if (line ~ /^kind:/)          { kind = trim(substr(line, 6)); in_meta = 0 }
    else if (line ~ /^metadata:/) { in_meta = 1 }
    else if (line ~ /^rules:/)    { in_rules = 1; in_meta = 0 }
    else                          { in_meta = 0 }
    print line
    return
  }
  if (in_meta && line ~ /^  name:/) {
    name = trim(substr(line, index(line, ":") + 1))
    print line
    return
  }
  if (in_rules) {
    if (line ~ /^  - apiGroups:/ || line ~ /^  - nonResourceURLs:/) {
      flush_rule()
      rulen = 0
      split("", rule_api); split("", rule_res); split("", res_line); split("", res_indent)
      rulebuf[++rulen] = line
      ruleIndent = indent_of(line)
      sub_state = (line ~ /apiGroups:/) ? "apigroups" : "nonres"
      return
    }
    if (rulen > 0) {
      if (line ~ /^    resources:/) { sub_state = "resources"; rulebuf[++rulen] = line; return }
      if (line ~ /^    verbs:/)     { sub_state = "verbs";     rulebuf[++rulen] = line; return }
      if (line ~ /^      - /) {
        item = trim(substr(line, index(line, "-") + 1))
        rulebuf[++rulen] = line
        if (sub_state == "apigroups")      rule_api[item] = 1
        else if (sub_state == "resources") { rule_res[item] = 1; res_line[item] = rulen; res_indent[item] = indent_of(line) }
        return
      }
      rulebuf[++rulen] = line
      return
    }
    print line
    return
  }
  print line
}

function flush_rule(   i, r, act, blockHelper, egHelper, line, ind, gi, res, emitted) {
  if (rulen == 0) return
  blockHelper = ""; egHelper = ""
  split("", gate)
  for (r = 1; r <= nrows; r++) {
    if (R_kind[r] != "" && R_kind[r] != "*" && R_kind[r] != kind) continue
    if (R_name[r] != "" && R_name[r] != "*" && R_name[r] != name) continue
    if (!api_in_rule(R_api[r])) continue
    act = R_action[r]
    if (act == "block") {
      if (res_in_rule(R_res[r])) blockHelper = R_helper[r]
    } else if (act == "resource") {
      if (R_res[r] in rule_res) gate[R_res[r]] = R_helper[r]
    } else if (act == "empty-guard") {
      if (resset_equals(R_res[r])) egHelper = R_helper[r]
    }
  }

  ind = ruleIndent
  if (egHelper != "")    print ind build_or(egHelper)
  if (blockHelper != "") print ind "{{- if eq (include \"" helperPrefix blockHelper "\" .) \"true\" }}"
  for (i = 1; i <= rulen; i++) {
    line = rulebuf[i]
    emitted = 0
    for (res in gate) {
      if (res_line[res] == i) {
        gi = res_indent[res]
        print gi "{{- if eq (include \"" helperPrefix gate[res] "\" .) \"true\" }}"
        print line
        print gi "{{- end }}"
        emitted = 1
        break
      }
    }
    if (!emitted) print line
  }
  if (blockHelper != "") print ind "{{- end }}"
  if (egHelper != "")    print ind "{{- end }}"
  rulen = 0
}

# Whole-file guard: nest an operandRequestsEnabled gate immediately inside the
# existing createRBAC wrapper. Idempotent: if the gate is already present the
# file is emitted unchanged.
function do_wholefile(   i, alreadyGuarded, ifIdx, endIdx) {
  alreadyGuarded = 0
  for (i = 1; i <= N; i++) if (L[i] ~ /include "ibm-licensing\.operandRequestsEnabled"/) { alreadyGuarded = 1; break }
  if (alreadyGuarded) { for (i = 1; i <= N; i++) print L[i]; return }

  ifIdx = 0; endIdx = 0
  for (i = 1; i <= N; i++) {
    if (L[i] ~ /createRBAC/ && L[i] ~ /{{- if/) ifIdx = i
    if (L[i] ~ /^{{- end }}[ \t]*$/)            endIdx = i
  }
  for (i = 1; i <= N; i++) {
    if (i == endIdx) print "{{- end }}"
    print L[i]
    if (i == ifIdx) print "{{- if eq (include \"" helperPrefix "operandRequestsEnabled\" .) \"true\" }}"
  }
}
