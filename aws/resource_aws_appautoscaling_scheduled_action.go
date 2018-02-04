package aws

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/applicationautoscaling"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
)

const awsAppautoscalingScheduleTimeLayout = "2006-01-02T15:04:05Z"

func resourceAwsAppautoscalingScheduledAction() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsAppautoscalingScheduledActionPut,
		Read:   resourceAwsAppautoscalingScheduledActionRead,
		Delete: resourceAwsAppautoscalingScheduledActionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"service_namespace": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"scalable_dimension": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"scalable_target_action": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"max_capacity": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
						"min_capacity": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
						},
					},
				},
			},
			"schedule": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"start_time": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"end_time": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"arn": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceAwsAppautoscalingScheduledActionPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appautoscalingconn

	input := &applicationautoscaling.PutScheduledActionInput{
		ScheduledActionName: aws.String(d.Get("name").(string)),
		ServiceNamespace:    aws.String(d.Get("service_namespace").(string)),
		ResourceId:          aws.String(d.Get("resource_id").(string)),
	}
	if v, ok := d.GetOk("scalable_dimension"); ok {
		input.ScalableDimension = aws.String(v.(string))
	}
	if v, ok := d.GetOk("schedule"); ok {
		input.Schedule = aws.String(v.(string))
	}
	if v, ok := d.GetOk("scalable_target_action"); ok {
		sta := &applicationautoscaling.ScalableTargetAction{}
		raw := v.([]interface{})[0].(map[string]interface{})
		if max, ok := raw["max_capacity"]; ok {
			sta.MaxCapacity = aws.Int64(int64(max.(int)))
		}
		if min, ok := raw["min_capacity"]; ok {
			sta.MinCapacity = aws.Int64(int64(min.(int)))
		}
		input.ScalableTargetAction = sta
	}
	if v, ok := d.GetOk("start_time"); ok {
		t, err := time.Parse(awsAppautoscalingScheduleTimeLayout, v.(string))
		if err != nil {
			return fmt.Errorf("Error Parsing Appautoscaling Scheduled Action Start Time: %s", err.Error())
		}
		input.StartTime = aws.Time(t)
	}
	if v, ok := d.GetOk("end_time"); ok {
		t, err := time.Parse(awsAppautoscalingScheduleTimeLayout, v.(string))
		if err != nil {
			return fmt.Errorf("Error Parsing Appautoscaling Scheduled Action End Time: %s", err.Error())
		}
		input.EndTime = aws.Time(t)
	}

	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.PutScheduledAction(input)
		if err != nil {
			if isAWSErr(err, applicationautoscaling.ErrCodeObjectNotFoundException, "") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if err != nil {
		return err
	}

	d.SetId(d.Get("name").(string) + "-" + d.Get("service_namespace").(string) + "-" + d.Get("resource_id").(string))
	return resourceAwsAppautoscalingScheduledActionRead(d, meta)
}

func resourceAwsAppautoscalingScheduledActionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appautoscalingconn

	name, serviceNamespace, _, err := decodeAppautoscalingScheduledActionID(d.Id())
	if err != nil {
		return err
	}
	input := &applicationautoscaling.DescribeScheduledActionsInput{
		ScheduledActionNames: []*string{aws.String(name)},
		ServiceNamespace:     aws.String(serviceNamespace),
	}
	resp, err := conn.DescribeScheduledActions(input)
	if err != nil {
		return err
	}
	if len(resp.ScheduledActions) < 1 {
		log.Printf("[WARN] Application Autoscaling Scheduled Action (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if len(resp.ScheduledActions) != 1 {
		return fmt.Errorf("Expected 1 scheduled action under %s, found %d", name, len(resp.ScheduledActions))
	}
	if *resp.ScheduledActions[0].ScheduledActionName != name {
		return fmt.Errorf("Scheduled Action (%s) not found", name)
	}

	action := resp.ScheduledActions[0]
	d.Set("name", action.ScheduledActionName)
	d.Set("service_namespace", action.ServiceNamespace)
	d.Set("resource_id", action.ResourceId)
	d.Set("scalable_dimension", action.ScalableDimension)
	d.Set("schedule", action.Schedule)
	if action.StartTime != nil {
		d.Set("start_time", action.StartTime.Format(awsAppautoscalingScheduleTimeLayout))
	}
	if action.EndTime != nil {
		d.Set("end_time", action.EndTime.Format(awsAppautoscalingScheduleTimeLayout))
	}
	if v := action.ScalableTargetAction; v != nil {
		sta := make(map[string]interface{}, 1)
		if v.MaxCapacity != nil {
			sta["max_capacity"] = *v.MaxCapacity
		}
		if v.MinCapacity != nil {
			sta["min_capacity"] = *v.MinCapacity
		}
		d.Set("scalable_target_action", []map[string]interface{}{sta})
	}
	d.Set("arn", action.ScheduledActionARN)
	return nil
}

func resourceAwsAppautoscalingScheduledActionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).appautoscalingconn

	name, serviceNamespace, resourceID, err := decodeAppautoscalingScheduledActionID(d.Id())
	if err != nil {
		return err
	}
	input := &applicationautoscaling.DeleteScheduledActionInput{
		ScheduledActionName: aws.String(name),
		ServiceNamespace:    aws.String(serviceNamespace),
		ResourceId:          aws.String(resourceID),
	}
	if v, ok := d.GetOk("scalable_dimension"); ok {
		input.ScalableDimension = aws.String(v.(string))
	}
	_, err = conn.DeleteScheduledAction(input)
	if err != nil {
		if isAWSErr(err, applicationautoscaling.ErrCodeObjectNotFoundException, "") {
			log.Printf("[WARN] Application Autoscaling Scheduled Action (%s) already gone, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}
	d.SetId("")
	return nil
}

func decodeAppautoscalingScheduledActionID(id string) (name, serviceNamespace, resourceID string, err error) {
	parts := strings.SplitN(id, "-", 2)
	if len(parts) != 3 {
		err = fmt.Errorf("Appautoscaling ScheduledAction ID must be of the form <Name>-<ServiceNamespace>-<ResourceID>, was provided: %s", id)
		return
	}
	name = parts[0]
	serviceNamespace = parts[1]
	resourceID = parts[2]
	return
}
