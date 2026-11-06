# Prototype for ADR-0018 / roadmap M23, not the shipped filter. Scored 9.4/10 with
# `sbomqs score --profile bsi-v2.1` on syft-cyclonedx.json (heraut, darwin/arm64).
# Known gaps: drops declared licences (original licences 0.2/10); owner values are hardcoded.
#
# Inputs: $zips (JSON array of {name, version, sha512} — the SHA-512 of each module zip in
# $(go env GOMODCACHE)/cache/download/<escaped path>/@v/<version>.zip, "" when absent),
# $ver (release tag), $file (asset name), $binsha (SHA-512 of the binary).
#
#   jq --argjson zips "$(cat zips.json)" --arg ver v0.82.1 \
#      --arg file heraut_0.82.1_darwin_arm64 --arg binsha "$(shasum -a 512 heraut | cut -d' ' -f1)" \
#      -f enrich.jq syft-cyclonedx.json
def owner: "Adaouat (https://github.com/adaouat)";
def bsi(f;e;a;s): [{name:"bsi:component:filename",value:f},{name:"bsi:component:executable",value:e},{name:"bsi:component:archive",value:a},{name:"bsi:component:structured",value:s}];
def lic: [.licenses[]? | .license.acknowledgement="concluded" | del(.license.url)];
($zips | map({key:(.name+"@"+.version), value:.sha512}) | from_entries) as $h
| (.components[] | select(.type=="library" and .name=="github.com/adaouat/heraut")) as $main
| .metadata.authors=[{name:"Adaouat",email:null}|del(.email)|.["bom-ref"]="author"] 
| .metadata.manufacturer={name:"Adaouat",url:["https://github.com/adaouat"]}
| .metadata.supplier={name:"Adaouat",url:["https://github.com/adaouat"]}
| .metadata.component = {
    "bom-ref":$main["bom-ref"], type:"application", name:"heraut", version:$ver,
    supplier:{name:"Adaouat",url:["https://github.com/adaouat"]},
    licenses:[{license:{id:"MIT",acknowledgement:"declared"}},{license:{id:"MIT",acknowledgement:"concluded"}}],
    purl:("pkg:golang/github.com/adaouat/heraut@"+$ver),
    externalReferences:[
      {type:"vcs",url:"https://github.com/adaouat/heraut"},
      {type:"distribution",url:("https://github.com/adaouat/heraut/releases/download/"+$ver+"/"+$file),hashes:[{alg:"SHA-512",content:$binsha}]},
      {type:"security-contact",url:"https://github.com/adaouat/heraut/security"}],
    properties:bsi($file;"executable";"no archive";"unstructured")}
| .components |= [ .[] | select(.type=="library" and .name!="github.com/adaouat/heraut")
    | ("https://" + (.name|sub("/v[0-9]+$";""))) as $src
    | .supplier={name:(.name|split("/")[0:2]|join("/")), url:[$src]}
    | .licenses = lic
    | .externalReferences = [
        {type:"vcs",url:$src},
        {type:"distribution",url:("https://proxy.golang.org/"+.name+"/@v/"+.version+".zip")}
          + (if $h[.name+"@"+.version] != "" then {hashes:[{alg:"SHA-512",content:$h[.name+"@"+.version]}]} else {} end)]
    | .properties = bsi(.version+".zip";"non-executable";"archive";"structured") ]
