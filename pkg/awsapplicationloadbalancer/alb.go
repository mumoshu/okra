package awsapplicationloadbalancer

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"golang.org/x/xerrors"
)

const (
	DefaultPriority = 100
)

func Delete(d *SyncInput) error {
	sess := d.Session

	sess.Config.Endpoint = &d.Address

	svc := elbv2.New(sess)

	listenerARN := d.Spec.ListenerARN

	describeRulesOutput, err := svc.DescribeRules(&elbv2.DescribeRulesInput{
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
	for _, r := range describeRulesOutput.Rules {
		if r.Priority != nil && *r.Priority == priorityStr {
			rule = r
		}
	}

	if rule != nil {
		input := &elbv2.DeleteRuleInput{RuleArn: rule.RuleArn}
		if res, err := svc.DeleteRule(input); err != nil {
			var appendix string

			if res != nil {
				appendix = fmt.Sprintf("\nOUTPUT:\n%v", *res)
			}

			log.Printf("Error: deleting rule: %w\nINPUT:\n%v%s", err, *input, appendix)

			return fmt.Errorf("deleting rule: %w", err)
		}
	}

	return nil
}
