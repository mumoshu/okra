package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"time"

	"github.com/mumoshu/okra/pkg/awsclicompat"
	"github.com/spf13/cobra"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

func main() {
	cmd := &cobra.Command{
		Use:   "aws",
		Short: "partial implementation of aws-cli in Go that has only `eks get-token` sub-command implemented",
	}

	eksCmd := &cobra.Command{
		Use: "eks",
	}

	var (
		clusterName, roleARN, region string
	)

	getTokenCmd := &cobra.Command{
		Use: "get-token",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := eksGetToken(os.Stdout, clusterName, roleARN, region)
			return err
		},
	}
	getTokenCmd.Flags().StringVar(&clusterName, "cluster-name", "", "Specify the name of the Amazon EKS  cluster to create a token for.")
	getTokenCmd.Flags().StringVar(&roleARN, "role", "", "Assume this role for credentials when signing the token.")

	cmd.PersistentFlags().StringVar(&region, "region", "", "The region to use. Overrides config/env settings.")

	eksCmd.AddCommand(getTokenCmd)
	cmd.AddCommand(eksCmd)

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

// This replicates the behavior of `aws eks get-token --cluster-name $CLUSTER_NAME`
// by using aws-iam-authenticator, which is the original implementation of the token generator that is,
// AFAIK, later ported to aws-cli.
//
// See https://github.com/aws/aws-cli/blob/develop/awscli/customizations/eks/get_token.py for the aws-cli implementation
// for reference.
func eksGetToken(out io.Writer, clusterID, roleARAN, region string) error {
	sess := awsclicompat.NewSession(region, "")

	gen, err := token.NewGenerator(true, false)
	if err != nil {
		return err
	}

	tok, err := gen.GetWithRoleForSession(clusterID, roleARAN, sess)
	if err != nil {
		return err
	}

	var execCredential struct {
		Kind       string            `json:"kind"`
		APIVersion string            `json:"apiVersion"`
		Spec       map[string]string `json:"spec"`
		Status     struct {
			ExpirationTimestamp string `json:"expirationTimestamp"`
			Token               string `json:"token"`
		} `json:"status"`
	}

	execCredential.Kind = "ExecCredential"
	execCredential.APIVersion = "client.authentication.k8s.io/v1alpha1"
	execCredential.Spec = map[string]string{}
	execCredential.Status.Token = tok.Token
	execCredential.Status.ExpirationTimestamp = tok.Expiration.Format(time.RFC3339)

	enc := json.NewEncoder(out)
	if err := enc.Encode(execCredential); err != nil {
		return err
	}

	return nil
}
