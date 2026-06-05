# Changelog

## [0.7.1](https://github.com/adaouat/forge/compare/v0.7.0..v0.7.1) - 2026-06-05

### 📚 Documentation

- *(roadmap)* Mark M7 theme palette + app wiring done - ([789a3d7](https://github.com/adaouat/forge/commit/789a3d73639c6e024cd5044f6a1bd906c9835c1e)) by @bchatard


### ⚙️ Miscellaneous Tasks

- Add release workflow (heraut-driven) - ([27fef85](https://github.com/adaouat/forge/commit/27fef851feeea9b71d4b64167c5ee37f7b0628c6)) by @bchatard

## [0.7.0](https://github.com/adaouat/forge/compare/v0.6.2..v0.7.0) - 2026-06-04

### 🚀 Features

- *(ui)* Add Palette for the shared family theme - ([4b19ee8](https://github.com/adaouat/forge/commit/4b19ee889014aa0f81690a356135b88ba468970e)) by @bchatard


### 📚 Documentation

- *(adr)* Add 0007 public API surface and stability contract - ([edaa2f2](https://github.com/adaouat/forge/commit/edaa2f2da825c4bf0239a656e8c99a350dbeb449)) by @bchatard

- *(adr)* Add 0008 family UI theme palette - ([271e450](https://github.com/adaouat/forge/commit/271e4505b497b7acbd9e7111e8662dd2b0649985)) by @bchatard

- *(guides)* Add Tier-2 sync guide and complete the M0-M6 roadmap - ([8af0d8f](https://github.com/adaouat/forge/commit/8af0d8fcb7f4f339d4e910e598f11745d33978ca)) by @bchatard

- *(roadmap)* Mark apps-depend-on-published-forge done - ([6981a90](https://github.com/adaouat/forge/commit/6981a902c72889aa2e6e4f6cee64b4ce19302143)) by @bchatard

- Update status from pre-implementation to shipped - ([93cc898](https://github.com/adaouat/forge/commit/93cc898609594daa361319be498900964fb8763e)) by @bchatard

## [0.6.2](https://github.com/adaouat/forge/compare/v0.6.1..v0.6.2) - 2026-06-04

### 📚 Documentation

- *(guides)* Update distribution guide for Homebrew casks - ([7bc8d2b](https://github.com/adaouat/forge/commit/7bc8d2b0cd73f2452e68eed842f7b591d64cd87e)) by @bchatard

- *(roadmap)* Mark Homebrew tap + casks done - ([9ac0ad5](https://github.com/adaouat/forge/commit/9ac0ad5d401e79adb369f30f0ffaa3006bc180df)) by @bchatard

## [0.6.1](https://github.com/adaouat/forge/compare/v0.6.0..v0.6.1) - 2026-06-04

### 💼 Other

- Bump Go to 1.26.4 for stdlib security fixes - ([6230dc8](https://github.com/adaouat/forge/commit/6230dc820cc6acd0cb74b03cdb31a333d85f9bb3)) by @bchatard


### 📚 Documentation

- *(roadmap)* Mark shared lint/CI workflow done - ([539959e](https://github.com/adaouat/forge/commit/539959e20dc393b02192069959b809b8e93a9405)) by @bchatard


### ⚙️ Miscellaneous Tasks

- Make go-ci coverage-threshold a required input - ([1bde7d9](https://github.com/adaouat/forge/commit/1bde7d93158afe08bf587118220c5912d9e69d87)) by @bchatard

## [0.6.0](https://github.com/adaouat/forge/compare/v0.5.0..v0.6.0) - 2026-06-04

### 🚀 Features

- *(updatecheck)* Add update check, install detection, and hint - ([da5cc1e](https://github.com/adaouat/forge/commit/da5cc1e8e68a5a3826475722d800b8298756d8cb)) by @bchatard


### 📚 Documentation

- *(adr)* Add 0005 updates via package managers - ([534d525](https://github.com/adaouat/forge/commit/534d525d353ac5a6dff19f75ffc6addfa5b47fbf)) by @bchatard

- *(adr)* Add 0006 shared CI reusable workflow - ([040cac3](https://github.com/adaouat/forge/commit/040cac34d5070c29d8224c6db79b5c0ce04e0292)) by @bchatard

- *(guides)* Add distribution guide and goreleaser sample - ([ca55570](https://github.com/adaouat/forge/commit/ca55570a1822a8de021c100117d43aa0771dd849)) by @bchatard

- *(roadmap)* Mark M5 heraut migration done - ([ce52ee6](https://github.com/adaouat/forge/commit/ce52ee6df398ff296a25f54a0080f869fa4a0f4d)) by @bchatard

- *(roadmap)* Mark M5.3 bifrost updatecheck wiring done - ([fd53477](https://github.com/adaouat/forge/commit/fd534778185ee7db3dc299554884009538c5623b)) by @bchatard

- *(roadmap)* Record forge as host for the shared lint/CI workflow - ([c891c05](https://github.com/adaouat/forge/commit/c891c05c34197fa2cdce0f6d0dd95335a324dd44)) by @bchatard

- *(roadmap)* Mark bifrost goreleaser convergence done - ([ecfa62f](https://github.com/adaouat/forge/commit/ecfa62fe8357342695a0a3f10524ff08972bfcfa)) by @bchatard

- *(roadmap)* Correct goreleaser name_template finding - ([352624d](https://github.com/adaouat/forge/commit/352624d993aaf255b94e11863747ef825b35d0cb)) by @bchatard


### ⚙️ Miscellaneous Tasks

- Add reusable lint/test workflow for the CLI family - ([759f467](https://github.com/adaouat/forge/commit/759f467c1e3ba7cae33a1646b84558907d498088)) by @bchatard

## [0.5.0](https://github.com/adaouat/forge/compare/v0.4.0..v0.5.0) - 2026-06-04

### 🚀 Features

- *(config)* Add strict YAML loader - ([72d3aa3](https://github.com/adaouat/forge/commit/72d3aa3161892e47bf2790325a5b69e1341af4db)) by @bchatard

- *(config)* Add app-parameterized path resolver - ([efed909](https://github.com/adaouat/forge/commit/efed9099daa051a904b6e49b4cba3c10fc9d56c7)) by @bchatard

- *(config)* Add ValidationError and ValidationErrors - ([12c7fd8](https://github.com/adaouat/forge/commit/12c7fd81dcfbf9c15dc59d9ea07cd36c2f8947c5)) by @bchatard


### 🚜 Refactor

- *(config)* Own the loader error wording - ([a029ca2](https://github.com/adaouat/forge/commit/a029ca2f8ef3bb8771101349d25cbb495a3c04ed)) by @bchatard


### 📚 Documentation

- *(roadmap)* Mark M4 config migrations done, close M4 - ([07749ca](https://github.com/adaouat/forge/commit/07749cac4f6bec085be927937f28f795612e2ca8)) by @bchatard

## [0.4.0](https://github.com/adaouat/forge/compare/v0.3.0..v0.4.0) - 2026-06-03

### 🚀 Features

- *(ui)* Add color/TTY detection and status helpers - ([339fad6](https://github.com/adaouat/forge/commit/339fad61a14409d178f0abd25fd17adef7c4feca)) by @bchatard

- *(ui)* Add output Mode value type - ([7523fbc](https://github.com/adaouat/forge/commit/7523fbcb403038d2ac361a844459d7ca06bfc78d)) by @bchatard

- *(ui)* Add header renderers - ([fe1c919](https://github.com/adaouat/forge/commit/fe1c9195b372e035331f85db435a331c3588e914)) by @bchatard

- *(ui)* Add Spinner task runner - ([faa497a](https://github.com/adaouat/forge/commit/faa497a937accecb5c7175f6cbc2b1539d0d5ad6)) by @bchatard

- *(ui)* Add Spinner.Step for completed numbered steps - ([18152ef](https://github.com/adaouat/forge/commit/18152efd5665e448d261349e240a16443a03cd72)) by @bchatard


### 📚 Documentation

- *(adr)* Add 0004 ui spinner task runner - ([d5493c2](https://github.com/adaouat/forge/commit/d5493c2c19ec48720e49ba1bfcaec4bbff3818b0)) by @bchatard

- *(roadmap)* Mark heraut/bifrost ui migration done, close M3 - ([3cf25e2](https://github.com/adaouat/forge/commit/3cf25e2ca47c2c964a982f14079f94f46e549923)) by @bchatard

- *(roadmap)* Mark M3.6 spinner migration done - ([36b965e](https://github.com/adaouat/forge/commit/36b965e3a4300200118c174994ee57f3bee42f06)) by @bchatard

- *(roadmap)* Mark bifrost numbered deploy steps done - ([8372162](https://github.com/adaouat/forge/commit/8372162186be92ae525b6fa492537f00c87489c9)) by @bchatard

## [0.3.0](https://github.com/adaouat/forge/compare/v0.2.0..v0.3.0) - 2026-06-03

### 🚀 Features

- *(exitcode)* Add ExitError, Wrap, and Resolve - ([f6e4316](https://github.com/adaouat/forge/commit/f6e43166754b41b88a2f00b12768502d9e800fc1)) by @bchatard

- *(exitcode)* Add shared exit-code constants - ([88cb144](https://github.com/adaouat/forge/commit/88cb1448f660878665e308a088b2ad5a4ac4fa63)) by @bchatard


### 📚 Documentation

- *(adr)* Add 0003 shared exit-code vocabulary - ([ffc3704](https://github.com/adaouat/forge/commit/ffc370435dbb55b26a6e1473e336dd766b154d8c)) by @bchatard

- *(roadmap)* Mark M2 heraut migration done - ([a90ace0](https://github.com/adaouat/forge/commit/a90ace076d4bd10b808d083fe3b95b1d285f86b7)) by @bchatard

- *(roadmap)* Mark M2 bifrost migration done, close M2 - ([60dfe21](https://github.com/adaouat/forge/commit/60dfe21155ba8fb6383443a1fbcac673bb7b9e23)) by @bchatard

- *(roadmap)* Mark heraut/bifrost exit-code adoption done - ([526e6a2](https://github.com/adaouat/forge/commit/526e6a2f6216604c164bada6f115da0fad2983a2)) by @bchatard

## [0.2.0](https://github.com/adaouat/forge/compare/v0.1.0..v0.2.0) - 2026-06-02

### 🚀 Features

- *(exec)* Add Runner interface and CmdRunner - ([37e02a0](https://github.com/adaouat/forge/commit/37e02a0e593e860485b4bed1ef06d674bea0df4c)) by @bchatard

- *(exec)* Add exectest MockRunner and FakeBin - ([2c724e7](https://github.com/adaouat/forge/commit/2c724e726eb0034fb5066d151b6bcef5c84d7b50)) by @bchatard

- *(exec)* Add RunDir for per-command working directory - ([f27dc81](https://github.com/adaouat/forge/commit/f27dc81b46097eeb21299c32326883c0ce81d549)) by @bchatard


### 📚 Documentation

- *(adr)* Add 0002 exec runner working directory - ([e9a7fd4](https://github.com/adaouat/forge/commit/e9a7fd4bab1e77d6b077e0b9b09e59b3af8c4bcf)) by @bchatard

- *(roadmap)* Mark M1 heraut wiring done - ([7dddb78](https://github.com/adaouat/forge/commit/7dddb78060158e2548fb611e9b110f9ca507a438)) by @bchatard

- *(roadmap)* Mark M1 bifrost wiring done, close M1 - ([146f8e8](https://github.com/adaouat/forge/commit/146f8e8751724bcc88f3b83d0ffadc3904bde4b8)) by @bchatard

- *(roadmap)* Note allow_fail stderr fix in M1.4 - ([7a3e910](https://github.com/adaouat/forge/commit/7a3e910469d4e1b8deed42d9ec4c1cda0f6a0cc4)) by @bchatard

## [0.1.0] - 2026-06-02

### 💼 Other

- Initialize the go module github.com/adaouat/forge - ([a2829bd](https://github.com/adaouat/forge/commit/a2829bd081d279e23b89593bb1d8837f2c741b39)) by @bchatard


### 📚 Documentation

- *(adr)* Mark ADR-0001 as accepted - ([cb29d5d](https://github.com/adaouat/forge/commit/cb29d5d36ad2a5d92d97b6de201e4ec331e7abb1)) by @bchatard

- *(roadmap)* Resolve dependency baseline and package naming, close M0 - ([64741bc](https://github.com/adaouat/forge/commit/64741bc0438c965dfb3bc17d7d570d833f0eead0)) by @bchatard

- Add shared-core extraction plan (ADR-0001 + roadmap) - ([0672cb0](https://github.com/adaouat/forge/commit/0672cb071806696dd3c62ad812744e977f26087e)) by @bchatard

- Port canonical .claude/rules adapted for a library - ([401dd93](https://github.com/adaouat/forge/commit/401dd93aded5791df9f4b21fbddb8297f0f44088)) by @bchatard

- Name the shared library forge, resolving module path decision - ([e719ad7](https://github.com/adaouat/forge/commit/e719ad7be949a7193c4e8f18520b15b63316ce4c)) by @bchatard

- Add docs/ index READMEs and specs skeleton - ([85ff5f8](https://github.com/adaouat/forge/commit/85ff5f87ce4285cbf6938a433a0f5b48212101e0)) by @bchatard

- Relocate conventions from .claude/rules to docs/rules - ([0c837d0](https://github.com/adaouat/forge/commit/0c837d016e2faf717d0dfe1f11c3a14d0dd88311)) by @bchatard


### ⚙️ Miscellaneous Tasks

- *(config)* Adapt mise/hk tooling for a library - ([2a27f14](https://github.com/adaouat/forge/commit/2a27f14f44c298025b47a2846aa21d76ef9a429c)) by @bchatard

- Init Adaouat Core - ([83de74c](https://github.com/adaouat/forge/commit/83de74c917f951990ece969ec8e93f706a119d44)) by @bchatard

- Add lint/test/build workflow - ([3911d4e](https://github.com/adaouat/forge/commit/3911d4ec1992c299c293deb7720b4e239ba7e9cc)) by @bchatard

