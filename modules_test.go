package test

import (
	"fmt"
	"os"
	"testing"
	//"time"
	//"io/ioutil"
	"strings"

	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/hashicorp/hcl/v2/hclsimple"

	//"go.mercari.io/hcledit"
	"terratest/tf"
)

func generateTF(version string, tfFile string) string {
	tfDir, err := os.MkdirTemp("", "terraform")
	if err != nil {
		panic(err)
	}
	versionsTF := strings.Replace(tf.VersionsTF, "VERSION", version, 1)

	fmt.Println(tfDir)

	tfFileAll := versionsTF + "\n" + tfFile

	err = os.WriteFile(tfDir+"/main.tf", []byte(tfFileAll), 0644)
	if err != nil {
		panic(err)
	}
	return tfDir
}

func TestModules(t *testing.T) {
	t.Parallel()

	versions := []string{
		//"~> 3.0",
		"~> 4.0",
		"~> 5.0",
	}

	modules := map[string]string{
		"sg": tf.SgTF,
	}

	results := [][]string{
		{"Version", "Validate"},
	}

	provider := "registry.terraform.io/hashicorp/aws"

	for _, version := range versions {
		for m, content := range modules {
			t.Run("TestProviderVersion_"+strings.Replace(version, "~>", "", -1)+m, func(t *testing.T) {
				validateSuccess := true
				planSuccess := true
				applySuccess := true
				providerVersions := make(map[string]string)
				tfDir := generateTF(version, content)
				lockFilePath := tfDir + "/.terraform.lock.hcl"
				defer os.RemoveAll(tfDir)
				terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
					TerraformDir: tfDir,
					NoColor:      true,
				})
				terraform.Init(t, terraformOptions)
				providerVersions, _ = GetProviderVersions(lockFilePath)
				fmt.Printf("Provider versions: %v\n", providerVersions)
				_, err := terraform.ValidateE(t, terraformOptions)
				if err != nil {
					validateSuccess = false
				}
				_, err = terraform.PlanE(t, terraformOptions)
				if err != nil {
					planSuccess = false
				}
				_, err = terraform.ApplyE(t, terraformOptions)
				defer terraform.Destroy(t, terraformOptions)
				if err != nil {
					applySuccess = false
				}
				defer func() {
					providerVersion, exists := providerVersions[provider]
					if !exists {
						t.Fatalf("Error: provider version not found in providerVersions: %v", providerVersions)
					}
					addResult(&results, m, providerVersion, validateSuccess, planSuccess, applySuccess)
				}()

			})
		}
	}
	printMarkdownMatrix(results)
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

func addResult(results *[][]string, module string, version string, validate bool, plan bool, apply bool) {
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
	*results = append(*results, []string{module, version, validateResult, planResult, applyResult})
}

func printMarkdownMatrix(results [][]string) {
	fmt.Println("\n### Terraform Provider Test Results")
	fmt.Println("Module | Version | Validate| Plan | Apply")
	fmt.Println("-------|---------|---------|-------|------|")
	for _, row := range results[1:] {
		fmt.Printf("| %s | %s | %s | %s | %s |\n", row[0], row[1], row[2], row[3], row[4])
	}
}
