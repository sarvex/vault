package transit

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/vault/api"
	"github.com/hashicorp/vault/sdk/testing/stepwise"
	dockerDriver "github.com/hashicorp/vault/sdk/testing/stepwise/drivers/docker"
	"github.com/mitchellh/mapstructure"
	"github.com/y0ssar1an/q"
)

// TestBackend_basic_derived_docker is an example test using the Docker Driver
func TestBackend_basic_derived_docker(t *testing.T) {
	// decryptData := make(map[string]interface{})
	stepwise.Run(t, stepwise.Case{
		Driver: dockerDriver.NewDockerDriver("transit"),
		Steps: []stepwise.Step{
			testAccStepwiseListPolicy(t, "test", true),
			testAccStepwiseWritePolicy(t, "test", true),
			testAccStepwiseListPolicy(t, "test", false),
			// testAccStepReadPolicy(t, "test", false, true),
			// testAccStepEncryptContext(t, "test", testPlaintext, "my-cool-context", decryptData),
			// testAccStepDecrypt(t, "test", testPlaintext, decryptData),
			// testAccStepEnableDeletion(t, "test"),
			// testAccStepDeletePolicy(t, "test"),
			// testAccStepReadPolicy(t, "test", true, true),
		},
	})
}

func testAccStepwiseWritePolicy(t *testing.T, name string, derived bool) stepwise.Step {
	ts := stepwise.Step{
		// Operation: logical.UpdateOperation,
		Operation: stepwise.WriteOperation,
		Path:      "keys/" + name,
		Data: map[string]interface{}{
			"derived": derived,
		},
		Check: func(resp *api.Secret) error {
			q.Q("--> stepwise write policy check func")
			return nil
		},
	}
	if os.Getenv("TRANSIT_ACC_KEY_TYPE") == "CHACHA" {
		ts.Data["type"] = "chacha20-poly1305"
	}
	return ts
}

func testAccStepwiseListPolicy(t *testing.T, name string, expectNone bool) stepwise.Step {
	return stepwise.Step{
		Operation: stepwise.ListOperation,
		Path:      "keys",
		Check: func(resp *api.Secret) error {
			q.Q("--> stepwise list check func")
			q.Q("resp in check:", resp)
			if (resp == nil || len(resp.Data) == 0) && !expectNone {
				return fmt.Errorf("missing response")
			}
			if expectNone && resp != nil {
				return fmt.Errorf("response data when expecting none")
			}

			if expectNone && resp == nil {
				return nil
			}

			q.Q("--> --> stepwise checking keys list")
			var d struct {
				Keys []string `mapstructure:"keys"`
			}
			if err := mapstructure.Decode(resp.Data, &d); err != nil {
				return err
			}
			if len(d.Keys) > 0 && d.Keys[0] != name {
				return fmt.Errorf("bad name: %#v", d)
			}
			if len(d.Keys) != 1 {
				return fmt.Errorf("only 1 key expected, %d returned", len(d.Keys))
			}
			q.Q("--> --> stepwise does with check")
			return nil
		},
	}
}