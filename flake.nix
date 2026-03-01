{
  description = "anker - reconstruct your workday after the fact";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
    pre-commit-hooks = {
      url = "github:cachix/pre-commit-hooks.nix";
      inputs.nixpkgs.follows = "nixpkgs";
    };
  };

  outputs = {
    self,
    nixpkgs,
    flake-utils,
    pre-commit-hooks,
  }:
    flake-utils.lib.eachDefaultSystem (system: let
      pkgs = nixpkgs.legacyPackages.${system};

      version = self.shortRev or self.dirtyShortRev or "dev";
      commit = self.rev or "dirty";
      date = self.lastModifiedDate or "unknown";

      ldflags = [
        "-s"
        "-w"
        "-X main.version=${version}"
        "-X main.commit=${commit}"
        "-X main.date=${date}"
      ];
    in {
      packages.default = pkgs.buildGoModule {
        pname = "anker";
        inherit version;
        src = ./.;
        vendorHash = "sha256-komX1AmHt2NoF1x6xsNa2RFkfVzOXfYEMPhT0zwMxjw=";
        env.CGO_ENABLED = 0;
        inherit ldflags;
        # tests run separately via checks.tests (they need git)
        doCheck = false;
      };

      checks = let
        pkg = self.packages.${system}.default;
      in {
        build = pkg;

        tests = pkg.overrideAttrs (old: {
          pname = "anker-tests";
          doCheck = true;
          nativeCheckInputs = [pkgs.git];
        });

        lint = pkgs.stdenvNoCC.mkDerivation {
          name = "anker-lint";
          src = ./.;
          nativeBuildInputs = [pkgs.go pkgs.golangci-lint];
          buildPhase = ''
            export HOME=$TMPDIR
            export CGO_ENABLED=0
            export GOLANGCI_LINT_CACHE=$TMPDIR/.golangci-lint
            export GOFLAGS="-mod=vendor"
            cp -r ${pkg.goModules} vendor
            golangci-lint run --timeout=5m
          '';
          installPhase = "touch $out";
        };

        pre-commit = pre-commit-hooks.lib.${system}.run {
          src = ./.;
          hooks = {
            gofmt.enable = true;
            # govet runs via golangci-lint in checks.lint (needs vendored modules)
            typos = {
              enable = true;
              settings.configPath = ".typos.toml";
            };
            alejandra.enable = true;
          };
        };
      };

      devShells.default = pkgs.mkShell {
        buildInputs = [
          pkgs.go
          pkgs.gopls
          pkgs.golangci-lint
        ];
        shellHook = self.checks.${system}.pre-commit.shellHook;
      };

      formatter = pkgs.alejandra;
    });
}
