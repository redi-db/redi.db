package updates

import _ "embed"

var VersionPath = "https://raw.githubusercontent.com/redi-db/redi.db/main/version.txt"

//go:embed ..\..\version.txt
var VERSION string
