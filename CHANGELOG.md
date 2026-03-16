# Changelog

## 1.0.0 (2026-03-16)


### Features

* **cli:** add ci summary output mode ([e55ad36](https://github.com/UreaLaden/log-sage/commit/e55ad367198f7d02099991aa0ca9bb21cdc7b58b))
* **cli:** add json output mode for analyze and ci ([c1ec3a8](https://github.com/UreaLaden/log-sage/commit/c1ec3a848d6834d22a80a6295d68ca42d0a4f11d))
* **cli:** add quiet sumamry output mode ([968b7e1](https://github.com/UreaLaden/log-sage/commit/968b7e1ffe509dfd553a08925e55cac556f87b09))
* **cli:** add stdin support to analyze command ([5ef8c7b](https://github.com/UreaLaden/log-sage/commit/5ef8c7b2e3a310f1ff63515c4a2f2144483f728a))
* **cli:** add top-n output truncation ([62d4a8d](https://github.com/UreaLaden/log-sage/commit/62d4a8d89c90415e3b1ca09bc206b1209b0e6818))
* **cli:** polich human-readable terminal output ([285d8b9](https://github.com/UreaLaden/log-sage/commit/285d8b948995d78fcea86413ff63a7678a01ea70))
* **detection:** add canonical issue registry ([0ce9a52](https://github.com/UreaLaden/log-sage/commit/0ce9a5207d2dc72f7a91119646204fa6144e625c))
* **detection:** add default detector wrapper ([349bcc8](https://github.com/UreaLaden/log-sage/commit/349bcc812ec2c46c3f7cd5eec64a9b65915765ed))
* **detection:** add default detector wrapper ([43dcfaf](https://github.com/UreaLaden/log-sage/commit/43dcfaf2511b72cf9b88a6b25136beea34cf959b))
* **detection:** add detector interface ([dce903e](https://github.com/UreaLaden/log-sage/commit/dce903e2921b724e307500ba059dacc785f91db4))
* **engine:** add default Analyze pipeline stub and constructor ([6f549e5](https://github.com/UreaLaden/log-sage/commit/6f549e5208f45570abd91f1d7566c22643918846))
* **engine:** wire extraction detection and scoring pipeline ([860ff4b](https://github.com/UreaLaden/log-sage/commit/860ff4b84055f963f9e8d7d0d4c8dba74c92dd58))
* **extraction:** add deterministic signal aggregation ([d95726c](https://github.com/UreaLaden/log-sage/commit/d95726c0ba17ec33c6b96a5dc55d988bcbf07006))
* **extraction:** add deterministic signal pattern extractor ([61fd719](https://github.com/UreaLaden/log-sage/commit/61fd719ff4505a1e8dc635f585134b35aab7e538))
* initialize Go module and scaffold Cobra CLI root ([007fccc](https://github.com/UreaLaden/log-sage/commit/007fccc2bb0b364c0d0c63dec2d5cfcd20056c9b))
* **k8s:** add kubectl adapter and real pod analysis command ([901ae64](https://github.com/UreaLaden/log-sage/commit/901ae643742f49a6f421d43835596ebaa2c82b26))
* **normalize:** add plaintext log parser ([4ee051f](https://github.com/UreaLaden/log-sage/commit/4ee051fe74872894dc5e8d00c048341104bc5550))
* **normalize:** implement multiline log entry grouping helper ([e616fd7](https://github.com/UreaLaden/log-sage/commit/e616fd753616c128edb85d40dbb57803a480bd06))
* **recommendation:** add command generation ([b41718c](https://github.com/UreaLaden/log-sage/commit/b41718c2409d2fb740511f690d5eaf17399ffab4))
* **recommendation:** add next-step generation ([36334f8](https://github.com/UreaLaden/log-sage/commit/36334f8f35d0955e10f037146bfb0bcff6e35498))
* **scoring:** add base candidate builder ([3a12139](https://github.com/UreaLaden/log-sage/commit/3a12139e62e6bc69975334a99b931cc704a0d1e4))
* **scoring:** add base candidate builder ([c610a01](https://github.com/UreaLaden/log-sage/commit/c610a0184e74b6512854f70da21011e322bac0f5))
* **scoring:** add candidate hypothesis contract ([8ea1e5b](https://github.com/UreaLaden/log-sage/commit/8ea1e5b534ad9d17c399806b7888292490908248))
* **scoring:** add candidate hypothesis contract ([c0a71c4](https://github.com/UreaLaden/log-sage/commit/c0a71c4ece26a05afd044502d985f8baa1e85d99))
* **scoring:** add confidence mapping pass ([37b68f7](https://github.com/UreaLaden/log-sage/commit/37b68f78b96c5d444f0bc79f49c66ad6c5e858c7))
* **scoring:** add failure phase inference ([140bb8f](https://github.com/UreaLaden/log-sage/commit/140bb8f088b2e856b0085746de017776e60be25e))
* **scoring:** add symptom relationship adjustment pass ([456aeac](https://github.com/UreaLaden/log-sage/commit/456aeac7f2595be75ecba622ba730302780a3873))
* **types:** add issue class contract ([175fe9e](https://github.com/UreaLaden/log-sage/commit/175fe9e1a0a6c362ba81d83162b72cd579914a4b))
* **types:** add signal extraction contracts ([4c2505f](https://github.com/UreaLaden/log-sage/commit/4c2505f9abb8866c18804e39c17a1d37e67b3736))


### Bug Fixes

* add missing CLI command files and anchor gitignore binary rule ([bb078d4](https://github.com/UreaLaden/log-sage/commit/bb078d47ec6098b8ae3c408f37183c651cf94d34))
* **normalizeJ:** allow larger logfmt scanner tokens ([42fd385](https://github.com/UreaLaden/log-sage/commit/42fd3851adcb9eef7236b5decd6134b42bf3cdb7))
* **normalize:** preserve exact json numeric field values ([770dc7a](https://github.com/UreaLaden/log-sage/commit/770dc7ac8b27841b5aa1ff59fcaecea0a45b6fe1))
* **release:** use PAT for release-please workflow ([bcf5653](https://github.com/UreaLaden/log-sage/commit/bcf5653a09a5ca052ba729ca8d0f92ea4108aaeb))
