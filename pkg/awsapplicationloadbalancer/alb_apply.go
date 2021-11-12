package awsapplicationloadbalancer

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/google/go-cmp/cmp"
	"github.com/mumoshu/okra/api/v1alpha1"
	"golang.org/x/xerrors"
)

func Sync(d SyncInput) error {
	log.SetFlags(log.Lshortfile)

	sess := d.Session

	sess.Config.Endpoint = &d.Address

	svc := elbv2.New(sess)

	listenerARN := d.Spec.ListenerARN
	lr := d.Spec.Listener.Rule
	destinations := d.Spec.Listener.Rule.Forward.TargetGroups
	desiredRuleActions := getRuleActions(destinations)
	desiredRuleConditions := getRuleConditions(lr)

	o, err := svc.DescribeRules(&elbv2.DescribeRulesInput{
		ListenerArn: aws.String(listenerARN),
	})
	if err != nil {
		return xerrors.Errorf("calling elbv2.DescribeRules: %w", err)
	}

	priority := d.Spec.Listener.Rule.Priority
	if priority == 0 {
		priority = DefaultPriority
	}
	priorityStr := strconv.Itoa(priority)

	var rule *elbv2.Rule
	for i := range o.Rules {
		r := o.Rules[i]

		if r.Priority != nil && *r.Priority == priorityStr {
			rule = r
		}
	}

	if rule == nil {
		log.Printf("Creating new rule for ALB listener %s", listenerARN)

		createRuleInput, err := ruleCreationInput(listenerARN, lr, destinations)
		o, err := svc.CreateRule(createRuleInput)
		if err != nil {
			return fmt.Errorf("creating listener rule: %w", err)
		}

		rule = o.Rules[0]

		log.Printf("Created new rule: %+v", *rule)
		return nil
	}

	log.Printf("Updating existing rule: %+v", *rule)

	var (
		conditionsModified bool
		actionsModified    bool
	)

	currentConditions := rule.Conditions

	for i := range rule.Conditions {
		// Otherwise we end up observing changes on Condition.Values even though
		// we can't set both Condition.Values and Condition.*.Values:
		//
		// alb_apply.go:83: Rule conditions has been changed: current (-), desired (+):
		//   []*elbv2.RuleCondition{
		//          &{
		//                  ... // 5 identical fields
		//                  QueryStringConfig: nil,
		//                  SourceIpConfig:    nil,
		// -                Values:            []*string{&"/*"},
		// +                Values:            nil,
		//          },
		//   }
		rule.Conditions[i].Values = nil
	}

	if d := cmp.Diff(currentConditions, desiredRuleConditions); d != "" {
		log.Printf("Rule conditions has been changed: current (-), desired (+):\n%s", d)

		conditionsModified = true
	}

	currentRuleActions := rule.Actions

	if d := cmp.Diff(currentRuleActions, desiredRuleActions); d != "" {
		log.Printf("Rule actions has been changed: current (-), desired (+):\n%s", d)

		actionsModified = true
	}

	if !conditionsModified && !actionsModified {
		return nil
	}

	log.Printf("Updating rule %s in-place, without traffic shifting", *rule.RuleArn)

	if len(desiredRuleConditions) == 0 {
		return errors.New("ALB does not support rule with no condition(s). Please specify one ore more from `hosts`, `path_patterns`, `methods`, `source_ips` and `headers`")
	}

	// ALB doesn't support traffic-weight between different rules.
	// We have no other way than modifying the rule in-place, which means no gradual traffic shiting is done.

	desiredActions := getRuleActions(destinations)
	modifyRuleInput := &elbv2.ModifyRuleInput{
		Actions:    desiredActions,
		Conditions: desiredRuleConditions,
		RuleArn:    rule.RuleArn,
	}

	if _, err := svc.ModifyRule(modifyRuleInput); err != nil {
		return fmt.Errorf("updating listener rule: %w", err)
	}

	return nil
}

func getRuleConditions(listenerRule v1alpha1.ListenerRule) []*elbv2.RuleCondition {
	// Create rule and set it to l.Rule
	ruleConditions := []*elbv2.RuleCondition{
		//	{
		//		Field:                   nil,
		//		HostHeaderConfig:        nil,
		//		HttpHeaderConfig:        nil,
		//		HttpRequestMethodConfig: nil,
		//		PathPatternConfig:       nil,
		//		QueryStringConfig:       nil,
		//		SourceIpConfig:          nil,
		//		Values:                  nil,
		//	}
	}

	// See this for how rule conditions should be composed:
	// https://cloudaffaire.com/aws-application-load-balancer-listener-rules-and-advance-routing-options
	// (I found it much readable and helpful than the official reference doc

	if len(listenerRule.Hosts) > 0 {
		ruleConditions = append(ruleConditions, &elbv2.RuleCondition{
			Field: aws.String("host-header"),
			HostHeaderConfig: &elbv2.HostHeaderConditionConfig{
				Values: aws.StringSlice(listenerRule.Hosts),
			},
		})
	}

	if len(listenerRule.PathPatterns) > 0 {
		ruleConditions = append(ruleConditions, &elbv2.RuleCondition{
			Field: aws.String("path-pattern"),
			PathPatternConfig: &elbv2.PathPatternConditionConfig{
				Values: aws.StringSlice(listenerRule.PathPatterns),
			},
		})
	}

	if len(listenerRule.Methods) > 0 {
		methods := make([]string, len(listenerRule.Methods))

		for i, m := range listenerRule.Methods {
			methods[i] = strings.ToUpper(m)
		}

		ruleConditions = append(ruleConditions, &elbv2.RuleCondition{
			Field: aws.String("http-request-method"),
			HttpRequestMethodConfig: &elbv2.HttpRequestMethodConditionConfig{
				Values: aws.StringSlice(methods),
			},
		})
	}

	if len(listenerRule.SourceIPs) > 0 {
		ruleConditions = append(ruleConditions, &elbv2.RuleCondition{
			Field: aws.String("source-ip"),
			SourceIpConfig: &elbv2.SourceIpConditionConfig{
				Values: aws.StringSlice(listenerRule.SourceIPs),
			},
		})
	}

	if len(listenerRule.Headers) > 0 {
		for name, values := range listenerRule.Headers {
			ruleConditions = append(ruleConditions, &elbv2.RuleCondition{
				Field: aws.String("http-header"),
				HttpHeaderConfig: &elbv2.HttpHeaderConditionConfig{
					HttpHeaderName: aws.String(name),
					Values:         aws.StringSlice(values),
				},
			})
		}
	}

	if len(listenerRule.QueryStrings) > 0 {
		var vs []*elbv2.QueryStringKeyValuePair

		for k, v := range listenerRule.QueryStrings {
			vs = append(vs, &elbv2.QueryStringKeyValuePair{
				Key:   aws.String(k),
				Value: aws.String(v),
			})
		}
		ruleConditions = append(ruleConditions, &elbv2.RuleCondition{
			Field: aws.String("query-string"),
			QueryStringConfig: &elbv2.QueryStringConditionConfig{
				Values: vs,
			},
		})
	}

	return ruleConditions
}

func getRuleActions(destinations []v1alpha1.ForwardTargetGroup) []*elbv2.Action {
	tgs := []*elbv2.TargetGroupTuple{}

	for _, d := range destinations {
		tgs = append(tgs, &elbv2.TargetGroupTuple{
			TargetGroupArn: aws.String(d.ARN),
			Weight:         aws.Int64(int64(d.Weight)),
		})
	}

	ruleActions := []*elbv2.Action{
		{
			ForwardConfig: &elbv2.ForwardActionConfig{
				TargetGroupStickinessConfig: nil,
				TargetGroups:                tgs,
			},
			Type: aws.String("forward"),
		},
	}

	return ruleActions
}

func ruleCreationInput(listenerARN string, listenerRule v1alpha1.ListenerRule, destinations []v1alpha1.ForwardTargetGroup) (*elbv2.CreateRuleInput, error) {
	ruleConditions := getRuleConditions(listenerRule)
	ruleActions := getRuleActions(destinations)

	createRuleInput := &elbv2.CreateRuleInput{
		Actions:     ruleActions,
		Priority:    aws.Int64(int64(listenerRule.Priority)),
		Conditions:  ruleConditions,
		ListenerArn: aws.String(listenerARN),
	}

	return createRuleInput, nil
}
