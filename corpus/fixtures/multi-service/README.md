# multi-service fixture

Two synthetic services (`api`, `billing`) plus two repo-level records, so
the `scope: repo` / `scope: service` split is at least expressible in a
real directory tree.

This fixture proves each record here is individually well-formed and
schema-valid. It does not prove that repo ∪ service union resolution
works: computing a service's effective record set is cross-file behaviour,
and no resolver exists at this milestone. The `deferred` entries (R-101,
R-103, R-108, R-109) mark exactly the tier-2 rules this fixture cannot
exercise yet.
