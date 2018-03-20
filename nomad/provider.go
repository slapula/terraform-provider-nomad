package nomad

import (
	"fmt"

	"github.com/hashicorp/nomad/api"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"github.com/hashicorp/vault/command/config"
)

type ProviderConfig struct {
	client     *api.Client
	vaultToken *string
}

func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"address": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOMAD_ADDR", nil),
				Description: "URL of the root of the target Nomad agent.",
			},

			"region": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOMAD_REGION", ""),
				Description: "Region of the target Nomad agent.",
			},
			"ca_file": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOMAD_CACERT", ""),
				Description: "A path to a PEM-encoded certificate authority used to verify the remote agent's certificate.",
			},
			"cert_file": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOMAD_CLIENT_CERT", ""),
				Description: "A path to a PEM-encoded certificate provided to the remote agent; requires use of key_file.",
			},
			"key_file": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOMAD_CLIENT_KEY", ""),
				Description: "A path to a PEM-encoded private key, required if cert_file is specified.",
			},
			"vault_token": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("VAULT_TOKEN", ""),
				Description: "Vault token if policies are specified in the job file.",
			},
			"secret_id": &schema.Schema{
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("NOMAD_TOKEN", ""),
				Description: "ACL token secret for API requests.",
			},
		},

		ConfigureFunc: providerConfigure,

		ResourcesMap: map[string]*schema.Resource{
			"nomad_acl_policy":          resourceACLPolicy(),
			"nomad_acl_token":           resourceACLToken(),
			"nomad_deployment":          resourceDeployment(),
			"nomad_job":                 resourceJob(),
			"nomad_quota_specification": resourceQuotaSpecification(),
		},
	}
}

// Get gets the value of the stored token, if any
func getToken() (string, error) {
	helper, err := config.DefaultTokenHelper()
	if err != nil {
		return "", fmt.Errorf("Error getting token helper: %s", err)
	}
	token, err := helper.Get()
	if err != nil {
		return "", fmt.Errorf("Error getting token: %s", err)
	}
	return token, nil
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	conf := api.DefaultConfig()
	conf.Address = d.Get("address").(string)
	conf.Region = d.Get("region").(string)
	conf.TLSConfig.CACert = d.Get("ca_file").(string)
	conf.TLSConfig.ClientCert = d.Get("cert_file").(string)
	conf.TLSConfig.ClientKey = d.Get("key_file").(string)
	conf.SecretID = d.Get("secret_id").(string)

	// Get the vault token from the conf, VAULT_TOKEN
	// or ~/.vault-token (in that order)
	var err error
	vaultToken := d.Get("vault_token").(string)
	if vaultToken == "" {
		vaultToken, err = getToken()
		if err != nil {
			return nil, err
		}
	}

	client, err := api.NewClient(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to configure Nomad API: %s", err)
	}

	res := ProviderConfig{
		client:     client,
		vaultToken: &vaultToken,
	}

	return res, nil
}
