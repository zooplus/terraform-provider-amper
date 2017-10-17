package provider

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/spirius/terraform-provider-amper/amper"
)

func dataSourceAmperContainer() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAmperContainerRead,

		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateName,
			},
			"attachment": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"policy_template_id": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateName,
						},
						"vars": {
							Type:     schema.TypeMap,
							Optional: true,
							Elem:     schema.TypeString,
						},
					},
				},
			},
			"policies": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func dataSourceAmperContainerRead(d *schema.ResourceData, meta interface{}) error {
	cc := meta.(*amper.Kernel)

	c, err := cc.NewContainer(d.Get("name").(string))

	if err != nil {
		return err
	}

	attachments := d.Get("attachment").(*schema.Set).List()

	for _, raw := range attachments {
		l := raw.(map[string]interface{})

		var vars = map[string]string{}

		if attr, ok := l["vars"]; ok {
			for k, v := range attr.(map[string]interface{}) {
				vars[k] = v.(string)
			}
		}

		_, err := c.AddAttachment(l["policy_template_id"].(string), l["account_name"].(string), vars)

		if err != nil {
			return err
		}
	}

	p, err, missing := c.Policy()

	if err != nil {
		return err
	}

	policyMap := map[string]string{}

	for account, policies := range p.AccountPolicies {
		for k, policy := range policies {
			s, err := json.MarshalIndent(policy, "", "  ")

			if err != nil {
				return err
			}

			policyMap[fmt.Sprintf("%s_%d", account, k)] = string(s)
		}

		policyMap[fmt.Sprintf("%s_count", account)] = fmt.Sprintf("%d", len(policies))
	}

	for _, a := range missing {
		log.Printf("[WARN] Policy template not found for '%s' in attachment '%s'", a, d.Id())
	}

	d.Set("policies", policyMap)

	d.SetId(d.Get("name").(string))

	return nil
}
