# multi-service fixture

Two synthetic services (`api`, `billing`) plus two repo-level records, so
the `scope: repo` / `scope: service` split is at least expressible in a
real directory tree.

This fixture proves both that each record here is individually well-formed
and schema-valid, and that R-101, R-103, R-108, R-109 and R-112 hold over
a real two-service tree. It still does not prove that a genuinely
multi-service *consumer* works. Only M4 shows that. jigctl's own
repository is single-service (`service_globs = []`), so the resolver's
two-tier behaviour is exercised only by this synthetic fixture.
