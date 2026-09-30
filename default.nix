{ lib, buildGoModule }:
buildGoModule {
  pname = "fleur";
  version = "0.0.1";

  src = ./.;

  vendorHash = "sha256-E0+30zRGmKdG5bkkGklLpzW7GAztm0iENYi34hn5hOU=";

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
