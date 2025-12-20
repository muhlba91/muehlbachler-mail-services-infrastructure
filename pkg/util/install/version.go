package install

import (
	"os"
	"strings"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert/yaml"
)

// Version reads the version of a service from a Docker Compose file.
// file: Path to the Docker Compose YAML file.
// service: Name of the service whose version is to be extracted.
// dockerComposeHash: Pulumi StringOutput representing the hash of the Docker Compose file.
func Version(file string, service string, dockerComposeHash pulumi.Output) pulumi.StringOutput {
	version, _ := dockerComposeHash.ApplyT(func(_ any) string {
		data, rErr := os.ReadFile(file)
		if rErr != nil {
			return ""
		}

		var parsed map[string]any
		if pErr := yaml.Unmarshal(data, &parsed); pErr != nil {
			return ""
		}

		v, ok := parsed["services"].(map[string]any)[service].(map[string]any)["image"].(string)
		if !ok {
			return ""
		}
		return strings.Split(v, ":")[1]
	}).(pulumi.StringOutput)

	return version
}
