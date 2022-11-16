{ lib, config, pkgs, ... }:

with lib;

let
  cfg = config.services.alertmanager-ntfy;
  settingsFormat = pkgs.formats.yaml { };
in {
  options.services.alertmanager-ntfy = {
    enable = mkEnableOption "Alertmanager notifications forwarder for ntfy.sh";
    settings = mkOption {
      type = types.submodule { freeformType = settingsFormat.type; };
      default = { };
      description = mdDoc ''
        Configuration for alertmanager-ntfy. Documented [here](https://github.com/alexbakker/alertmanager-ntfy).
      '';
    };
    extraSettingsPath = mkOption {
      type = types.nullOr types.path;
      default = null;
      description = ''
        Extra configuration file (YAML) to load and merge into the config generated from the settings defined by the NixOS module.
        This can be used to pass credentials so that they don't end up in the Nix store.
      '';
    };
  };

  config = let
    configuration = settingsFormat.generate "settings.yml" cfg.settings;
  in mkIf cfg.enable {
    systemd.services.alertmanager-ntfy = {
      enable = true;
      description = "Alertmanager notifications forwarder for ntfy.sh";
      after = [ "network.target" ];
      wantedBy = [ "multi-user.target" ];
      serviceConfig = {
        Type = "simple";
        DynamicUser = true;
        LoadCredential = lib.mkIf (cfg.extraSettingsPath != null) "auth.yml:${cfg.extraSettingsPath}";
        ExecStart = "${pkgs.alertmanager-ntfy}/bin/alertmanager-ntfy --configs ${configuration}${lib.optionalString (cfg.extraSettingsPath != null) ",\"\${CREDENTIALS_DIRECTORY}/auth.yml\""}";
        Restart = "always";
        RestartSec = 5;

        UMask = "077";
        NoNewPrivileges = true;
        ProtectSystem = "strict";
        ProtectHome = true;
        PrivateTmp = true;
        PrivateDevices = true;
        PrivateUsers = true;
        ProtectHostname = true;
        ProtectClock = true;
        ProtectKernelTunables = true;
        ProtectKernelModules = true;
        ProtectKernelLogs = true;
        ProtectControlGroups = true;
        ProtectProc = "invisible";
        ProcSubset = "pid";
        RestrictAddressFamilies = [ "AF_INET" "AF_INET6" ];
        RestrictNamespaces = true;
        LockPersonality = true;
        MemoryDenyWriteExecute = true;
        RestrictRealtime = true;
        RestrictSUIDSGID = true;
        RemoveIPC = true;
        PrivateMounts = true;
        SystemCallArchitectures = "native";
        SystemCallFilter = ["@system-service" "~@privileged" ];
        CapabilityBoundingSet = null;
      };
    };
  };
}
