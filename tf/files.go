package tf

import (
	_ "embed"
)

//go:embed common/versions.tf
var VersionsTF string

//go:embed sg/main.tf
var SgTF string
