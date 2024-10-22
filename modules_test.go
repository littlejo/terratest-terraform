package test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/gruntwork-io/terratest/modules/terraform"
	terratest "github.com/gruntwork-io/terratest/modules/testing"
	"github.com/hashicorp/hcl/v2/hclsimple"

	"terratest/tf"
)

func generateTF(version string, moduleVersion string, tfFile string) string {
	tfDir, err := os.MkdirTemp("", "terraform")
	if err != nil {
		panic(err)
	}
	versionsTF := strings.Replace(tf.VersionsTF, "VERSION", version, 1)
	mainTF := strings.Replace(tfFile, "VERSION", moduleVersion, 1)

	fmt.Println(tfDir)

	tfFileAll := versionsTF + "\n" + mainTF

	err = os.WriteFile(tfDir+"/main.tf", []byte(tfFileAll), 0644)
	if err != nil {
		panic(err)
	}
	return tfDir
}

func TestModules(t *testing.T) {

	versions := []string{
		//"~> 3.0",
		"~> 4.0",
		"~> 5.0",
	}

	type Module struct {
		Content  string
		Versions []string
	}

	modules := map[string]Module{
		"sg": {
			Content:  tf.SgTF,
			Versions: tf.SgTFVersion,
		},
		"secretsmanager": {
			Content:  tf.SecretsTF,
			Versions: tf.SecretsTFVersion,
		},
	}

	for _, version := range versions {
		for name, module := range modules {
			for _, moduleVersion := range module.Versions {
				t.Run("TestProviderVersion_"+strings.Replace(version, "~>", "", -1)+name+moduleVersion, func(t *testing.T) {
					t.Parallel()
					testModule(t, version, module.Content, moduleVersion, name)
				})
			}
		}
	}

}

func testModule(t *testing.T, version, content, moduleVersion string, moduleName string) {
	provider := "registry.terraform.io/hashicorp/aws"
	providerVersions := make(map[string]string)
	tfDir := generateTF(version, moduleVersion, content)
	lockFilePath := tfDir + "/.terraform.lock.hcl"
	defer os.RemoveAll(tfDir)
	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: tfDir,
		NoColor:      true,
	})
	terraform.Init(t, terraformOptions)
	providerVersions, _ = GetProviderVersions(lockFilePath)
	fmt.Printf("Provider versions: %v\n", providerVersions)
	validateSuccess := executeTerraformStep(t, terraform.ValidateE, terraformOptions, "validate")
	planSuccess := executeTerraformStep(t, terraform.PlanE, terraformOptions, "plan")
	applySuccess := executeTerraformStep(t, terraform.ApplyE, terraformOptions, "apply")
	defer terraform.Destroy(t, terraformOptions)
	defer func() {
		providerVersion, exists := providerVersions[provider]
		if !exists {
			t.Fatalf("Error: provider version not found in providerVersions: %v", providerVersions)
		}
		addResult(moduleName, moduleVersion, providerVersion, validateSuccess, planSuccess, applySuccess)
	}()
}

func executeTerraformStep(t *testing.T, step func(t terratest.TestingT, options *terraform.Options) (string, error), terraformOptions *terraform.Options, stepName string) bool {
	_, err := step(t, terraformOptions)
	if err != nil {
		t.Logf("%s failed: %v", stepName, err)
		return false
	}
	return true
}

type LockFile struct {
	Provider []ProviderInfo `hcl:"provider,block"`
}

type ProviderInfo struct {
	Name        string   `hcl:"name,label"`
	Version     string   `hcl:"version"`
	Constraints string   `hcl:"constraints,optional"`
	Hashes      []string `hcl:"hashes,optional"`
}

func GetProviderVersions(lockFilePath string) (map[string]string, error) {
	var lockFile LockFile

	// Decode the terraform.lock.hcl file
	err := hclsimple.DecodeFile(lockFilePath, nil, &lockFile)
	if err != nil {
		return nil, fmt.Errorf("failed to parse lock file: %w", err)
	}

	// Create a map to store the providers and their versions
	providerVersions := make(map[string]string)

	// Populate the map with provider name as key and version as value
	for _, provider := range lockFile.Provider {
		fmt.Printf("Found provider: %s, version: %s\n", provider.Name, provider.Version)
		providerVersions[provider.Name] = provider.Version
	}

	if len(providerVersions) == 0 {
		fmt.Println("Warning: No provider versions found in lock file.")
	}

	return providerVersions, nil
}

func addResult(module string, moduleVersion string, version string, validate bool, plan bool, apply bool) {
	validateResult := "success"
	if !validate {
		validateResult = "fail"
	}
	planResult := "success"
	if !plan {
		planResult = "fail"
	}
	applyResult := "success"
	if !apply {
		applyResult = "fail"
	}
	results := [][]string{
		{module, moduleVersion, version, validateResult, planResult, applyResult},
	}
	printMarkdownMatrix(results, "/tmp/results")
}

func printMarkdownMatrix(results [][]string, filename string) error {
	file, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("could not stat file: %w", err)
	}

	if fileInfo.Size() == 0 {
		_, err = file.WriteString("\n### Terraform Provider Test Results\n")
		if err != nil {
			return fmt.Errorf("could not write to file: %w", err)
		}

		_, err = file.WriteString("Module | module Version | provider Version | Validate| Plan | Apply\n")
		if err != nil {
			return fmt.Errorf("could not write to file: %w", err)
		}

		_, err = file.WriteString("-------|---------|---------|-------|------|-------|\n")
		if err != nil {
			return fmt.Errorf("could not write to file: %w", err)
		}
	}

	for _, row := range results {
		_, err := file.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n", row[0], row[1], row[2], row[3], row[4], row[5]))
		if err != nil {
			return fmt.Errorf("could not write to file: %w", err)
		}
	}

	return nil
}
