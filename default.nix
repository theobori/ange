{ lib, buildGoModule }:
buildGoModule {
  pname = "fleur";
  version = "0.0.1";

  src = ./.;

  vendorHash = "sha256-3Glc/5N4/Liq//AHJHwMNbRABoPfjDZxqXZqt6AG0CA=";

  ldflags = [
    "-s"
    "-w"
  ];

  meta = {
    description = "Gopher webring based on fleur ";
    homepage = "https://github.com/theobori/ange";
    license = lib.licenses.mit;
    mainProgram = "fleur";
  };
}
