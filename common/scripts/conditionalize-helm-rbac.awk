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
  # Top-level {{- if ... }} / {{- end }} are balanced through a stack so the outer
  # createRBAC wrapper survives while object guards (whose openers carry an
  # ibm-licensing helper include) are removed. Indented rule/item guards are dropped
  # by the per-line filters; a templated subject (name: {{ include ... }}) is content,
  # not a {{- if opener, so it is preserved.
  cn = 0; sp = 0
  for (i = 1; i <= N; i++) {
    s = L[i]
    if (s ~ /^{{- if/) {
      if (s ~ /include "ibm-licensing\./) { stack[++sp] = "STRIP" }
      else                                { stack[++sp] = "KEEP"; C[++cn] = s }
      continue
    }
    if (s ~ /^{{- end }}[ \t]*$/) {
      if (sp > 0) { if (stack[sp--] == "KEEP") C[++cn] = s; continue }
      C[++cn] = s; continue
    }
    if (s ~ /{{- if .*include "ibm-licensing\./) continue
    if (s ~ /^[ \t]+{{- end }}[ \t]*$/)          continue
    C[++cn] = s
  }

  build_object_guards()

  kind = ""; name = ""; in_meta = 0; in_rules = 0; sub_state = ""; rulen = 0
  objPending = ""
  for (i = 1; i <= cn; i++) process(C[i], i)
  flush_rule()
  close_block()
  if (objPending != "") print "{{- end }}"
}

# For each document boundary (---) in the stripped stream, record the object-guard
# helper (if any) that should wrap the document it opens, keyed by line index.
function build_object_guards(   i, j, dk, dn, im) {
  for (i = 1; i <= cn; i++) openAt[i] = ""
  for (i = 1; i <= cn; i++) {
    if (C[i] !~ /^---/) continue
    dk = ""; dn = ""; im = 0
    for (j = i + 1; j <= cn && C[j] !~ /^---/; j++) {
      if (C[j] ~ /^kind:/)              { dk = trim(substr(C[j], 6)); im = 0 }
      else if (C[j] ~ /^metadata:/)     { im = 1 }
      else if (C[j] ~ /^[A-Za-z]/)      { im = 0 }
      else if (im && C[j] ~ /^  name:/) { dn = trim(substr(C[j], index(C[j], ":") + 1)) }
    }
    openAt[i] = object_helper(dk, dn)
  }
}

# Object rows match on (kind,name) only; apiGroup/resources are "*".
function object_helper(k, n,   r) {
  for (r = 1; r <= nrows; r++) {
    if (R_action[r] != "object") continue
    if (R_kind[r] != "" && R_kind[r] != "*" && R_kind[r] != k) continue
    if (R_name[r] != "" && R_name[r] != "*" && R_name[r] != n) continue
    return R_helper[r]
  }
  return ""
}

function process(line, idx,   item) {
  if (line ~ /^---/) {
    flush_rule()
    close_block()
    if (objPending != "") { print "{{- end }}"; objPending = "" }
    print line
    kind = ""; name = ""; in_meta = 0; in_rules = 0; sub_state = ""
    if (openAt[idx] != "") {
      print "{{- if eq (include \"" helperPrefix openAt[idx] "\" .) \"true\" }}"
      objPending = openAt[idx]
    }
    return
  }
  # Top-level {{- end }} (the createRBAC closer): close any open object guard first
  # so the object {{- end }} nests inside the createRBAC wrapper.
  if (line ~ /^{{- end }}[ \t]*$/) {
    flush_rule()
    close_block()
    if (objPending != "") { print "{{- end }}"; objPending = "" }
    print line
    return
  }
  if (line ~ /^[A-Za-z]/) {
    flush_rule()
    close_block()
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
  # cluster-monitoring-view subject follows the active operand SA. The
  # templated form is content (not a {{- if opener), so the strip phase keeps it and
  # this rewrite is a no-op on re-run.
  if (name == "ibm-license-service-cluster-monitoring-view" && line ~ /^    name: ibm-license-service$/) {
    print "    name: {{ include \"" helperPrefix "operandServiceAccount\" . }}"
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

  # Coalesce adjacent block guards that share the same helper into one wrapper: keep
  # the guard open across consecutive matching rules and close it only when the helper
  # changes or at a rule-section boundary (close_block, from the document handlers).
  # This avoids emitting a separate {{- if <helper> }} ... {{- end }} per rule when
  # several neighbours gate on the same flag (e.g. kubeRBACAuthEnabled wrapping both
  # tokenreviews and subjectaccessreviews).
  if (openBlock != "" && openBlock != blockHelper) { print openBlockIndent "{{- end }}"; openBlock = "" }
  if (blockHelper != "" && openBlock == "") {
    print ind "{{- if eq (include \"" helperPrefix blockHelper "\" .) \"true\" }}"
    openBlock = blockHelper; openBlockIndent = ind
  }

  if (egHelper != "")    print ind build_or(egHelper)
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
  if (egHelper != "")    print ind "{{- end }}"
  # The block guard close is deferred to close_block() so adjacent same-helper rules share it.
  rulen = 0
}

# Close the currently open coalesced block guard, if any. Called at every rule-section
# boundary (document separator, top-level key, createRBAC closer, EOF).
function close_block() {
  if (openBlock != "") { print openBlockIndent "{{- end }}"; openBlock = "" }
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
