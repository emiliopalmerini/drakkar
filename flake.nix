{
  description = "Drakkar — OpenViking MCP bridge";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
      in
      {
        packages.default = pkgs.buildGoModule {
          pname = "drakkar";
          version = "0.1.0";
          src = ./.;
          vendorHash = "sha256-yBZ2PLxiI7GYAqLM52yV5yq7dVWXQcBqn66l6BgqJn8=";
        };

        devShells.default = pkgs.mkShell {
          packages = [ pkgs.go ];
        };
      }
    );
}
