package nomad

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
)

func resourceDeployment() *schema.Resource {
	return &schema.Resource{
		Create: resourceDeploymentUpdate,
		Update: resourceDeploymentUpdate,
		Delete: resourceDeploymentUpdate,
		Read:   resourceDeploymentRead,
		Exists: resourceDeploymentExists,

		Schema: map[string]*schema.Schema{
			"id": {
				Description: "Deployment ID.",
				Required:    true,
				Type:        schema.TypeString,
			},

			"state": {
				Description: "Deployment State.",
				Required:    true,
				Type:        schema.TypeString,
			},

			"namepsace": {
				Description: "Deployment State.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"job_id": {
				Description: "Job ID.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},

			"job_version": {
				Description: "Job Version.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},

			"job_modify_index": {
				Description: "Job Modify Index.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},

			"job_create_index": {
				Description: "Job Create Index.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},

			"task_groups": {
				Description: "Task Groups.",
				Optional:    true,
				Computed:    true,
				Type:        schema.TypeList,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"placed_canaries": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"auto_revert": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"promoted": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"desired_canaries": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"desired_total": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"placed_alloc": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"healthy_alloc": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"unhealthy_alloc": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},

			"status": {
				Description: "Deployment Status.",
				Required:    true,
				Type:        schema.TypeString,
			},

			"status_description": {
				Description: "Deployment Status Description.",
				Required:    true,
				Type:        schema.TypeString,
			},

			"modify_index": {
				Description: "Modify Index.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},

			"create_index": {
				Description: "Create Index.",
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceDeploymentUpdate(d *schema.ResourceData, meta interface{}) error {
	providerConfig := meta.(ProviderConfig)
	client := providerConfig.client

	id := d.Id()
	deploymentState := d.Get("state").(string)

	if deploymentState == "fail" {
		log.Printf("[DEBUG] Manually failing deployment: %q", id)
		_, _, err := client.Deployments().Fail(id, nil)
		if err != nil {
			return fmt.Errorf("error applying deployment update: %s", err)
		}
	}

	if deploymentState == "pause" {
		log.Printf("[DEBUG] Pausing deployment: %q", id)
		_, _, err := client.Deployments().Pause(id, true, nil)
		if err != nil {
			return fmt.Errorf("error applying deployment update: %s", err)
		}
	}

	if deploymentState == "promote" {
		log.Printf("[DEBUG] Promoting deployment: %q", id)
		_, _, err := client.Deployments().PromoteAll(id, nil)
		if err != nil {
			return fmt.Errorf("error applying deployment update: %s", err)
		}
	}

	if deploymentState == "resume" {
		log.Printf("[DEBUG] Resuming deployment: %q", id)
		_, _, err := client.Deployments().Pause(id, false, nil)
		if err != nil {
			return fmt.Errorf("error applying deployment update: %s", err)
		}
	}

	return nil
}

func resourceDeploymentExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	providerConfig := meta.(ProviderConfig)
	client := providerConfig.client

	id := d.Id()
	log.Printf("[DEBUG] Checking if deployment exists: %q", id)
	_, _, err := client.Deployments().Info(id, nil)
	if err != nil {
		// As of Nomad 0.4.1, the API client returns an error for 404
		// rather than a nil result, so we must check this way.
		if strings.Contains(err.Error(), "404") {
			return false, nil
		}

		return false, fmt.Errorf("error checking for deployment: %#v", err)
	}

	return true, nil
}

func resourceDeploymentRead(d *schema.ResourceData, meta interface{}) error {
	providerConfig := meta.(ProviderConfig)
	client := providerConfig.client

	id := d.Id()
	log.Printf("[DEBUG] Getting deployment status: %q", id)
	deployment, _, err := client.Deployments().Info(id, nil)
	if err != nil {
		// As of Nomad 0.4.1, the API client returns an error for 404
		// rather than a nil result, so we must check this way.
		if strings.Contains(err.Error(), "404") {
			return err
		}

		return fmt.Errorf("error checking for deployment: %#v", err)
	}

	d.Set("id", deployment.ID)
	d.Set("namespace", deployment.Namespace)
	d.Set("job_id", deployment.JobID)
	d.Set("job_version", deployment.JobVersion)
	d.Set("job_modify_index", deployment.JobModifyIndex)
	d.Set("job_create_index", deployment.JobCreateIndex)
	d.Set("task_groups", deployment.TaskGroups)
	d.Set("status", deployment.Status)
	d.Set("status_description", deployment.StatusDescription)
	d.Set("modify_index", deployment.ModifyIndex)
	d.Set("create_index", deployment.CreateIndex)

	return nil
}
