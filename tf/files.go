package tf

import (
	_ "embed"
)

//go:embed common/versions.tf
var VersionsTF string

//go:embed sg/main.tf
var SgTF string
var SgTFVersion = []string{"v0.2.0"}

//var SgTFVersion = []string{"v0.2.0", "v0.1.1"}

//go:embed secretsmanager/main.tf
var SecretsTF string
var SecretsTFVersion = []string{"v0.2.0"}

//var SecretsTFVersion = []string{"v0.2.0", "v0.1.2"}
