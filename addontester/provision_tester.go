package addontester

import (
	"encoding/json"
	"fmt"

	"github.com/bitrise-team/bitrise-add-on-testing-kit/addonprovisioner"
	"github.com/bitrise-team/bitrise-add-on-testing-kit/utils"
)

// ProvisionTesterParams ...
type ProvisionTesterParams struct {
	AppSlug   string
	APIToken  string
	Plan      string
	WithRetry bool
}

type env struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type provisionResp struct {
	Envs []env `json:"envs"`
}

// Provision ...
func (c *Tester) Provision(params ProvisionTesterParams, remainingRetries int) error {

	if len(params.AppSlug) == 0 {
		params.AppSlug, _ = utils.RandomHex(8)
	}

	if len(params.APIToken) == 0 {
		params.APIToken, _ = utils.RandomHex(8)
	}

	c.logger.Printf("\nProvisioning details:")
	c.logger.Printf("App slug: %s", params.AppSlug)
	c.logger.Printf("API token: %s", params.APIToken)
	c.logger.Printf("Plan: %s", params.Plan)
	c.logger.Printf("Should retry: %v", params.WithRetry)
	if params.WithRetry {
		c.logger.Printf("No. of test: %d.", numberOfTestsWithRetry-remainingRetries)
	}

	status, body, err := c.addonClient.Provision(addonprovisioner.ProvisionRequestParams{
		AppSlug:  params.AppSlug,
		APIToken: params.APIToken,
		Plan:     params.Plan,
	})

	if err != nil {
		return fmt.Errorf("Provisioning failed: %s", err)
	}

	c.logger.Printf("\nResponse status: %d", status)
	c.logger.Printf("Response body: %v\n", body)

	if status < 200 || status > 299 {
		return fmt.Errorf("Provisioning request resulted in a non-2xx response")
	}

	var pr provisionResp

	if err := json.Unmarshal([]byte(body), &pr); err != nil {
		return fmt.Errorf("JSON parsing of response failed: %s", err)
	}

	if len(pr.Envs) > 0 {
		for _, e := range pr.Envs {
			if len(e.Key) == 0 {
				return fmt.Errorf("ENV var key is not present: %v", e)
			}

			if len(e.Value) == 0 {
				return fmt.Errorf("ENV var value is not present: %v", e)
			}

			c.logger.Printf("ENV var processed succesfully: %s: %s", e.Key, e.Value)
		}
	} else {
		c.logger.Printf("No ENV vars to check in response")
	}

	c.logger.Println("\nProvisioning success.")

	if params.WithRetry && remainingRetries > 0 {
		return c.Provision(params, remainingRetries-1)
	}

	return nil
}
