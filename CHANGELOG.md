# Changelog

## [0.2.0](https://github.com/bbmumford/contracts/compare/v0.1.0...v0.2.0) (2026-11-25)


### Features

* auth.v1 carries a partition's shared sign-in settings (SignInSettings, Get/SetSignInSettings) ([4c3609f](https://github.com/bbmumford/contracts/commit/4c3609fb66dac2d2e4fbb88e8c01edbdeba9bf8c))
* CreateSubscriptionRequest carries idempotencyKey (field 6), the caller's key for the one create it retries ([02a8c1d](https://github.com/bbmumford/contracts/commit/02a8c1d0fcdf6bf06a7589f30414ba6f2225be96))
* partition_root.v1 carries the partition root's signing and anchor messages ([03cbeac](https://github.com/bbmumford/contracts/commit/03cbeac031e328bda93aa80a601c54f090cd8b11))
* ResolveSessionResponse.partitionVia, how the sign-in chose the session's partition ([6136884](https://github.com/bbmumford/contracts/commit/6136884446c71b2434f403e1b0eb118304b45338))
