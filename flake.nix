{
  inputs = {
    nix-go = {
      inputs.nixpkgs.follows = "nixpkgs";
      url = "github:matthewdargan/nix-go";
    };
    nixpkgs.url = "nixpkgs/nixos-unstable";
    parts.url = "github:hercules-ci/flake-parts";
    pre-commit-hooks = {
      inputs.nixpkgs.follows = "nixpkgs";
      url = "github:cachix/pre-commit-hooks.nix";
    };
  };
  outputs = inputs:
    inputs.parts.lib.mkFlake {inherit inputs;} {
      imports = [inputs.pre-commit-hooks.flakeModule];
      perSystem = {
        config,
        inputs',
        lib,
        pkgs,
        ...
      }: {
        devShells.default = pkgs.mkShell {
          packages = [inputs'.nix-go.packages.go inputs'.nix-go.packages.golangci-lint];
          shellHook = "${config.pre-commit.installationScript}";
        };
        packages.go = inputs'.nix-go.legacyPackages.buildGoModule {
          doCheck = false;
          meta = with lib; {
            description = "Packages and commands for using Plan 9 from Go";
            homepage = "https://github.com/matthewdargan/go";
            license = licenses.mit;
            maintainers = with maintainers; [matthewdargan];
          };
          pname = "go";
          postInstall = "rm $out/bin/acme";
          src = ./.;
          vendorHash = "sha256-qwoYzbfyek/472x24srND/9qX2UsiKzJleLV7cFDVsY=";
          version = "0.1.0";
        };
        pre-commit = {
          check.enable = false;
          settings = {
            hooks = {
              alejandra.enable = true;
              deadnix.enable = true;
              golangci-lint = {
                enable = true;
                package = inputs'.nix-go.packages.golangci-lint;
              };
              gotest = {
                enable = true;
                package = inputs'.nix-go.packages.go;
              };
              statix.enable = true;
            };
            src = ./.;
          };
        };
      };
      systems = ["aarch64-darwin" "aarch64-linux" "x86_64-darwin" "x86_64-linux"];
    };
}
