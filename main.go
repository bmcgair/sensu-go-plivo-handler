package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/plivo/plivo-go"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

//Declare our variable types here
var (
	accountSid string
	authToken  string
	fromNumber string
	recipient  string
	message  string
	stdin      *os.File
)

//Start main function
func main() {
	rootCmd := configureRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

//Configure the root command and add our flags
func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sensu-go-plivo-handler",
		Short: "The Sensu Go Handler for plivo",
		RunE:  run,
	}
	cmd.Flags().StringVarP(&accountSid,
		"accountSid",
		"s",
		os.Getenv("PLIVO_ACCOUNT_SID"),
		"The account SID for your plivo account, uses the environment variable PLIVO_ACCOUNT_SID by default")

	cmd.Flags().StringVarP(&authToken,
		"authToken",
		"t",
		os.Getenv("PLIVO_AUTH_TOKEN"),
		"The authorization token for your plivo account, uses the environment variable PLIVO_AUTH_TOKEN by default")

	cmd.Flags().StringVarP(&fromNumber,
		"fromNumber",
		"f",
		os.Getenv("PLIVO_FROM_NUMBER"),
		"Your plivo phone number, uses the environment variable PLIVO_FROM_NUMBER by default")

//	_ = cmd.MarkFlagRequired("fromNumber")

	cmd.Flags().StringVarP(&recipient,
		"recipient",
		"r",
		os.Getenv("PLIVO_RECIPIENT_LIST"),
		"The recipient's phone number, uses the environment variable PLIVO_RECIPIENT_LIST")

//  _ = cmd.MarkFlagRequired("recipient")

	return cmd
}

func run(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		_ = cmd.Help()
		return fmt.Errorf("invalid argument(s) received")
	}

	if stdin == nil {
		stdin = os.Stdin
	}

	eventJSON, err := ioutil.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %s", err)
	}

	event := &types.Event{}
	err = json.Unmarshal(eventJSON, event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stdin data: %s", err)
	}

	if err = event.Validate(); err != nil {
		return fmt.Errorf("failed to validate event: %s", err)
	}

	if !event.HasCheck() {
		return fmt.Errorf("event does not contain check")
	}

	return sendSMS(event)
}

//This function will send an SMS when receive an alert in error state.
func sendSMS(event *types.Event) error {

	//Set up our message we want to send
	message := "Sensu alert for " + event.Check.Name + " on " + event.Entity.Name + ". Check output: " + event.Check.Output

	//Set up a plivo client with our accountSid & authToken
	client, err := plivo.NewClient(accountSid, authToken, &plivo.ClientOptions{})

	if err != nil {
		panic(err)
	}

	//Send our message to our recipient
	client.Messages.Create(plivo.MessageCreateParams{
		Src:  fromNumber,
		Dst:  recipient,
		Text: message,
	})

	return nil
}
