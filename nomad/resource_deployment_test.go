package nomad

import (
	"fmt"
	"testing"

	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestCheckNomadDeployment_Basic(t *testing.T) {
	var deployment api.Deployment
	deploymentId := acctest.RandString(8)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testProviders,
		CheckDestroy: testCheckNomadDeploymentFail,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testCheckNomadDeploymentConfig_basic, deploymentId),
				Check: resource.ComposeTestCheckFunc(
					testCheckNomadDeploymentExists("nomad_deployment.foobar", &deployment),
					testCheckNomadDeploymentAttributes(&deployment, deploymentId),
					resource.TestCheckResourceAttr(
						"nomad_deployment.foobar", "id", deploymentId),
					resource.TestCheckResourceAttr(
						"nomad_deployment.foobar", "state", "resume"),
				),
			},
		},
	})
}

func testCheckNomadDeploymentFail(s *terraform.State) error {
	providerConfig := testProvider.Meta().(ProviderConfig)
	client := providerConfig.client

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "nomad_deployment" {
			continue
		}

		// Try to find the domain
		_, _, err := client.Deployments().Fail(rs.Primary.ID, nil)

		if err == nil {
			return fmt.Errorf("Deployment still running")
		}
	}

	return nil
}
func testCheckNomadDeploymentExists(a string, deployment *api.Deployment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[a]

		if !ok {
			return fmt.Errorf("Not found: %s", a)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Deployment ID is set")
		}

		providerConfig := testProvider.Meta().(ProviderConfig)
		client := providerConfig.client

		foundDeployment, _, err := client.Deployments().Info(rs.Primary.ID, nil)

		if err != nil {
			return err
		}

		if foundDeployment.ID != rs.Primary.ID {
			return fmt.Errorf("Deployment not found")
		}

		*deployment = *foundDeployment

		return nil
	}
}

func testCheckNomadDeploymentAttributes(deployment *api.Deployment, id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if deployment.ID != id {
			return fmt.Errorf("Bad Deployment ID: %s", deployment.ID)
		}

		return nil
	}
}

const testCheckNomadDeploymentConfig_basic = `
resource "nomad_deployment" "foobar" {
	id		= "%s"
	state	= "resume"
}`
