{ lib, buildGoModule }:
buildGoModule {
  pname = "fleur";
  version = "0.0.1";

  src = ./.;

  vendorHash = "sha256-Lu25iyDQe+puKTSKHxYF2tSbWQajGmI9AInZ7Y0ZSMI=";

  ldflags = [
    "-s"
    "-w"
  ];

  meta = {
    description = "Gopher webring based on fleur";
    homepage = "https://github.com/theobori/ange";
    license = lib.licenses.mit;
    mainProgram = "fleur";
  };
}
