package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceAwsMediaStoreCorsPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsMediaStoreCorsPolicyPut,
		Read:   resourceAwsMediaStoreCorsPolicyRead,
		Update: resourceAwsMediaStoreCorsPolicyPut,
		Delete: resourceAwsMediaStoreCorsPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"container_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"cors_policy": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 100,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allowed_headers": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allowed_methods": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"allowed_origins": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"expose_headers": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"max_age_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceAwsMediaStoreCorsPolicyPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	input := &mediastore.PutCorsPolicyInput{
		ContainerName: aws.String(d.Get("container_name").(string)),
		CorsPolicy:    expandMediaStoreCorsPolicy(d.Get("cors_policy").(*schema.Set).List()),
	}

	_, err := conn.PutCorsPolicy(input)
	if err != nil {
		return err
	}

	d.SetId(d.Get("container_name").(string))
	return resourceAwsMediaStoreCorsPolicyRead(d, meta)
}

func resourceAwsMediaStoreCorsPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	input := &mediastore.GetCorsPolicyInput{
		ContainerName: aws.String(d.Id()),
	}

	resp, err := conn.GetCorsPolicy(input)
	if err != nil {
		if isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
			log.Printf("[WARN] MediaStore Cors Policy %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		if isAWSErr(err, mediastore.ErrCodeCorsPolicyNotFoundException, "") {
			log.Printf("[WARN] MediaStore Cors Policy %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	d.Set("container_name", d.Id())
	d.Set("cors_policy", flattenMediaStoreCorsPolicy(resp.CorsPolicy))
	return nil
}

func resourceAwsMediaStoreCorsPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).mediastoreconn

	input := &mediastore.DeleteCorsPolicyInput{
		ContainerName: aws.String(d.Id()),
	}

	_, err := conn.DeleteCorsPolicy(input)
	if err != nil {
		if isAWSErr(err, mediastore.ErrCodeContainerNotFoundException, "") {
			return nil
		}
		if isAWSErr(err, mediastore.ErrCodeCorsPolicyNotFoundException, "") {
			return nil
		}
		return err
	}

	return nil
}

func expandMediaStoreCorsPolicy(configured []interface{}) []*mediastore.CorsRule {
	rules := make([]*mediastore.CorsRule, 0, len(configured))

	for _, raw := range configured {
		rule := &mediastore.CorsRule{}
		data := raw.(map[string]interface{})

		if v, ok := data["allowed_headers"]; ok {
			rule.AllowedHeaders = expandStringSet(v.(*schema.Set))
		}
		if v, ok := data["allowed_methods"]; ok {
			rule.AllowedMethods = expandStringSet(v.(*schema.Set))
		}
		if v, ok := data["allowed_origins"]; ok {
			rule.AllowedOrigins = expandStringSet(v.(*schema.Set))
		}
		if v, ok := data["expose_headers"]; ok {
			rule.ExposeHeaders = expandStringSet(v.(*schema.Set))
		}
		if v, ok := data["max_age_seconds"]; ok {
			rule.MaxAgeSeconds = aws.Int64(int64(v.(int)))
		}
		rules = append(rules, rule)
	}
	return rules
}

func flattenMediaStoreCorsPolicy(corsRules []*mediastore.CorsRule) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)

	for _, rule := range corsRules {
		m := make(map[string]interface{})
		if len(rule.AllowedHeaders) > 0 {
			m["allowed_headers"] = flattenStringList(rule.AllowedHeaders)
		}
		if len(rule.AllowedMethods) > 0 {
			m["allowed_methods"] = flattenStringList(rule.AllowedMethods)
		}
		if len(rule.AllowedOrigins) > 0 {
			m["allowed_origins"] = flattenStringList(rule.AllowedOrigins)
		}
		if len(rule.ExposeHeaders) > 0 {
			m["expose_headers"] = flattenStringList(rule.ExposeHeaders)
		}
		if rule.MaxAgeSeconds != nil {
			m["max_age_seconds"] = *rule.MaxAgeSeconds
		}
		result = append(result, m)
	}

	return result
}
