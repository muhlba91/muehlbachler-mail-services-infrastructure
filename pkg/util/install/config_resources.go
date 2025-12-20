package install

import "github.com/pulumi/pulumi/sdk/v3/go/pulumi"

// CollectResourceOptions collects Pulumi ResourceOptions from a list of Pulumi Outputs.
// resources: A slice of Pulumi Outputs representing resources.
func CollectResourceOptions(resources []pulumi.Output) []pulumi.ResourceOption {
	var resourceOptions []pulumi.ResourceOption
	for _, r := range resources {
		r.ApplyT(func(res any) error {
			resOpt, _ := res.(pulumi.ResourceOption)
			resourceOptions = append(resourceOptions, resOpt)
			return nil
		})
	}
	return resourceOptions
}
