{
  description = "Env sensor prometheus collector";

  # Nixpkgs / NixOS version to use.
  inputs.nixpkgs.url = "nixpkgs/nixos-23.05";

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
        let
          pkgs = nixpkgsFor.${system};
        in
        {
          go-home-sensors = pkgs.mkShell {
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

            vendorSha256 = "sha256-fSOV+xPugDwNOu+WHOdaOwUbo8YHbIE9jCutotaZkJY=";
          };
        });

      nixosModules = {
        home-sensors = { config, lib, pkgs, ... }:
          with lib;
          let
            cfg = config.alex-nt.services.home-sensors;
            toml = pkgs.formats.toml { };
            configFilePath = toml.generate "sensors.toml" cfg.settings;
          in
          {
            options.alex-nt.services.home-sensors = {
              enable = mkEnableOption "Enables the home-sensors exporter service";
              settings = {
                port = lib.mkOption {
                  type = lib.types.port;
                  default = 2112;
                  description = ''
                    Which port this service should listen on.
                  '';
                };
                frequency = lib.mkOption {
                  type = types.str;
                  default = "15s";
                  description = ''
                    Interval at which to refresh the metrics.
                  '';
                };
                sensors = builtins.mapAttrs
                  (name: value: {
                    enable = mkEnableOption "Enables ${s} sensor";
                    register = lib.mkOption {
                      type = types.number;
                      default = value;
                      description = ''
                        i2c register where the sensor is located.
                      '';
                    };
                  })
                  {
                    bme68x = 118; #0x76
                    scd4x = 98; #0x62
                    pmsa003i = 18; #0x12
                  };
              };
              exporters = {
                prometheus.enable = mkEnableOption "Enable prometheus exporter";
                sqlite = {
                  enable = mkEnableOption "Enable sqlite exporter";
                  db = lib.mkOption {
                    type = types.str;
                    default = "/var/lib/${pname}/metrics.db";
                    description = ''
                      database connection string.
                    '';
                  };
                };
              };
            };

            config = mkIf cfg.enable {
              systemd.services.alexnt-home-sensors = {
                wantedBy = [ "multi-user.target" ];

                serviceConfig =
                  let
                    pkg = self.packages.${pkgs.system}.default;
                  in
                  {
                    Restart = "on-failure";
                    ExecStart = "${pkg}/bin/go-home-sensors --config.file=${configFilePath}";
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
      defaultPackage = forAllSystems (system: self.packages.${system}.default);
    };
}
