## 1.0.0 (PoC release)

BREAKING CHANGES:
- `values` on `autodns_record` is now a Set instead of a List. DNS treats the
  values at a name and type as an unordered set, and AutoDNS returns them in a
  different order than they were written, which made every reordering show up
  as a permanent plan diff on records nobody had touched. Configuration is
  unchanged (`values = [...]` still works), but the state representation
  differs, so `terraform plan` should be reviewed carefully on first upgrade.

## 0.2.0 (PoC release)

Zone records are now fetched once per zone instead of once per record resource.
The AutoDNS API has no per-record read, so every `autodns_record` used to
download the whole zone and discard all but one record. Plans against a zone
with many records now make two API calls instead of hundreds.

IMPROVEMENTS:
- Cache zone records per provider process, invalidated whenever a record is
  created, updated or deleted.
- Reads no longer block on the mutex that serializes writes, and parallel reads
  of a cold zone collapse into a single request.

BUG FIXES:
- Return an error instead of panicking when the API returns no zone.
- Remove a record from state when it no longer exists in the zone, so the plan
  proposes recreating it rather than failing.
- Invalidate the per-zone cache before a mutation as well as after, so a
  snapshot cached earlier in the same run cannot be served stale after a write.
- Hold the write lock for reading while fetching a zone, so a read can no
  longer overlap an in-flight write. The previous single global lock gave this
  for free; splitting it for caching had removed it.

## 0.1.2 (PoC release)

Fixed an issue where the validation fails when the `values` property in `record_resource` is still unknown.

## 0.1.1 (PoC release)

Small wording fix.

## 0.1.0 (PoC release)

This release is meant as a proof of concept and implement the minimum amount of code to test the provider.

FEATURES:
- `autodns_record` resource.
- `autodns_zone` data source.