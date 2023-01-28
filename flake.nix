{
  description = "Env sensor prometheus collector";

  # Nixpkgs / NixOS version to use.
  inputs.nixpkgs.url = "nixpkgs/nixos-22.11";

  outputs = { self, nixpkgs }:
    let

      # to work with older version of flakes
      lastModifiedDate = self.lastModifiedDate or self.lastModified or "19700101";

      # Generate a user-friendly version number.
      version = builtins.substring 0 8 lastModifiedDate;

      # System types to support.
      supportedSystems = [ "x86_64-linux" "x86_64-darwin" "aarch64-linux" "aarch64-darwin" ];

      # Helper function to generate an attrset '{ x86_64-linux = f "x86_64-linux"; ... }'.
      forAllSystems = nixpkgs.lib.genAttrs supportedSystems;

      # Nixpkgs instantiated for supported system types.
      nixpkgsFor = forAllSystems (system: import nixpkgs { inherit system; });

    in
    {
      apps = forAllSystems (system: {
        default = {
          type = "app";
          program = "${self.packages.${system}.default}/bin/go-home-sensors";
        };
      });
      devShells = forAllSystems (system:
        let pkgs = nixpkgsFor.${system};
        in {
          default = pkgs.mkShell {
            buildInputs = with pkgs; [ go gopls gotools go-tools ];
          };
        });
      # Provide some binary packages for selected system types.
      packages = forAllSystems (system:
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          default = pkgs.buildGoModule {
            pname = "go-home-sensors";
            inherit version;
            # In 'nix develop', we don't need a copy of the source tree
            # in the Nix store.
            src = ./.;

            # This hash locks the dependencies of this package. It is
            # necessary because of how Go requires network access to resolve
            # VCS.  See https://www.tweag.io/blog/2021-03-04-gomod2nix/ for
            # details. Normally one can build with a fake sha256 and rely on native Go
            # mechanisms to tell you what the hash should be or determine what
            # it should be "out-of-band" with other tooling (eg. gomod2nix).
            # To begin with it is recommended to set this, but one must
            # remeber to bump this hash when your dependencies change.
            #vendorSha256 = pkgs.lib.fakeSha256;

            vendorSha256 = "sha256-bgShhrc+53HW3NLpC+2upi/atyZhwGghqijR3JHjIdo=";
          };
        });

      nixosModules = {
        home-sensors = { config, lib, pkgs, ... }:
          with lib;
          let cfg = config.alex-nt.services.home-sensors;
          in {
            options.alex-nt.services.home-sensors = {
              enable = mkEnableOption "Enables the home-sensors prometheus-collector service";
              port = lib.mkOption {
                type = lib.types.port;
                default = 2112;
                description = ''
                  Which port this service should listen on.
                '';
              };
              listenAddress = mkOption {
                type = types.str;
                default = "0.0.0.0";
                description = lib.mdDoc ''
                  Address to listen on for the exporter.
                '';
              };
            };

            config = mkIf cfg.enable {
              systemd.services.alexnt-home-sensors = {
                wantedBy = [ "multi-user.target" ];

                serviceConfig =
                  let pkg = self.packages.${pkgs.system}.default;
                  in {
                    Restart = "on-failure";
                    ExecStart = "${pkg}/bin/go-home-sensors --web.listen-address=${cfg.listenAddress}:${builtins.toString cfg.port}";
                    DynamicUser = "yes";
                    SupplementaryGroups = [ "i2c" ];
                  };
              };
            };
          };
      };

      # The default package for 'nix build'. This makes sense if the
      # flake provides only one package or there is a clear "main"
      # package.
      defaultPackage = forAllSystems (system: self.packages.${system}.go-home-sensors);
    };
}
